package estrategias

import (
	"math"
	"time"
	"log"
	"math/rand"

	pb "agentegabornelas/pb"
)

// Nodo representa un estado del tablero en nuestro árbol de búsqueda
type Nodo struct {
	Estado       		*pb.GameState 	// El estado del tablero en este nodo
	Padre       		*Nodo   	  	// El nodo padre
	Hijos     			[]*Nodo   		// Las posibles jugadas siguientes
	Jugada        		*pb.Point     	// La jugada que nos llevó a este estado
	Victorias      		float64      	// Cantidad de victorias simuladas desde aquí
	Visitas       		int           	// Cuántas veces hemos pasado por este nodo
	JugadasPendientes 	[]*pb.Point   	// Jugadas que aún no hemos explorado
	Completado 			bool			// Condicion de hijos explorados completamente
}

// Bot mantiene la memoria del árbol entre turnos
type Bot struct {
	Root *Nodo
}

// NuevoBot inicializa un bot vacío
func NuevoBot() *Bot {
	return &Bot{Root: nil}
}

// MCTS actualiza el árbol y ejecuta MCTS---------------------------------------------------------------------------------------------------
func (bot *Bot) MCTS(estadoActual *pb.GameState, miColor pb.PlayerColor, defensa bool, turnos int32) []*pb.Point {
	// Inicializar Raiz
	if bot.Root == nil {
		log.Println("Raiz inexistente")
		bot.Root = &Nodo{Estado: estadoActual, JugadasPendientes: espaciosLibres(estadoActual, miColor, defensa)}
	} else {
		// Buscamos si el estado actual ya fue evaluado y es un hijo de la raiz actual
		encontrado := false
		for _, hijo := range bot.Root.Hijos {
			// Verificamos si el hijo de nuestra raiz es igual al estado actual
			if esMismoTablero(hijo.Estado.Board, estadoActual.Board) {
				bot.Root = hijo
				bot.Root.Padre = nil // Podamos el árbol por encima de la nueva raíz
				encontrado = true
				log.Println("Hijo encontrado")
				break
			}
		}
		// Si el rival hizo un movimiento no explorado
		if !encontrado {
			log.Println("jugada no explorada")
			bot.Root = &Nodo{Estado: estadoActual, JugadasPendientes: espaciosLibres(estadoActual, miColor, defensa)}
		}
	}

	// MCTS
	tiempo := time.Now().Add(4500*time.Millisecond)
	for time.Now().Before(tiempo) {

		if bot.Root.Completado {
            break //Se completó el arbol, salimos del bucle antes
        }

		nodoActual := bot.Root
		colorActual := miColor // Empezamos asumiendo que es nuestro turno en la Raíz
		// Selección
		for len(nodoActual.JugadasPendientes) == 0 && len(nodoActual.Hijos) > 0 {
			nodoActual = seleccion(nodoActual)
			if turnos % 2 == 1 {
				colorActual = colorOponente(colorActual) // Cambiamos de turno al bajar de nivel cuando queda 1 turno
				turnos++
			}else{turnos--}
		}

		resultado := 0.0
		// Expansión
		if len(nodoActual.JugadasPendientes) > 0 {
			nodoActual = expandir(nodoActual, colorActual, defensa, turnos)
			if turnos % 2 == 1 {
				colorActual = colorOponente(colorActual) // Cambiamos de turno al expandir cuando queda 1 turno
				turnos++
			}else{turnos--}
	
			// Simulación
			resultado = simularPartida(nodoActual.Estado, miColor, colorActual, defensa, turnos)
		}else{
			// No hay jugadas pendientes y no hay hijos. Es un estado terminal. Lo marcamos como completado.
            nodoActual.Completado = true
            
            // Como ya es el final, 'simularPartida' simplemente leerá quién ganó sin hacer movimientos aleatorios
            resultado = simularPartida(nodoActual.Estado, miColor, colorActual, defensa, turnos)
		}
		// Retropropagación
		retropropagar(nodoActual, resultado)
	}
	
	// Elegir el mejor movimiento (heuristica)
	mejorHijo := seleccionarHijoMasVisitado(bot.Root)
	// Actualizamos la raíz a la jugada que vamos a hacer
	bot.Root = mejorHijo
	bot.Root.Padre = nil
	return []*pb.Point{mejorHijo.Jugada}
}

// Selección--------------------------------------------------------------------------------------------------------------------------------
func seleccion(nodo *Nodo) *Nodo {
	// Aquí se aplica la fórmula matemática UCT (Upper Confidence Bound Applied to Trees)
	// Balancea la "Explotación" (jugar lo que sabemos que es bueno) 
	// con la "Exploración" (probar movimientos nuevos).
	var mejorHijo *Nodo
	mejorPuntaje := math.Inf(-1)
	
	for _, hijo := range nodo.Hijos {
		// Fórmula UCB1 estándar: (Victorias / Visitas) + C * sqrt(ln(VisitasPadre) / VisitasHijo)
		// C suele ser 1.41 (Raíz de 2)
		puntaje := (hijo.Victorias / float64(hijo.Visitas)) + math.Sqrt(2)*math.Sqrt(math.Log(float64(nodo.Visitas))/float64(hijo.Visitas))
		if puntaje > mejorPuntaje {
			mejorPuntaje = puntaje
			mejorHijo = hijo
		}
	}
	return mejorHijo
}

// Expansión--------------------------------------------------------------------------------------------------------------------------------
func expandir(nodo *Nodo, colorToca pb.PlayerColor, defensa bool, turnos int32) *Nodo {
	// Sacar un movimiento al azar de los pendientes
	indice := rand.Intn(len(nodo.JugadasPendientes))
	movimiento := nodo.JugadasPendientes[indice]

	// Eliminar el movimiento de la lista (Swap and pop rápido)
	nodo.JugadasPendientes[indice] = nodo.JugadasPendientes[len(nodo.JugadasPendientes)-1]
	nodo.JugadasPendientes = nodo.JugadasPendientes[:len(nodo.JugadasPendientes)-1]

	// Clonar el estado y aplicar el movimiento con el color recibido
	nuevoEstado := clonarEstado(nodo.Estado)
	jugar(nuevoEstado, movimiento, colorToca)

	if turnos % 2 == 1 {
		colorToca = colorOponente(colorToca) // Cambiamos de turno al expandir cuando queda 1 turno
	} 

	// Crear el nuevo nodo
	nuevoNodo := &Nodo{
		Estado:        		nuevoEstado,
		Padre:       		nodo,
		Jugada:         	movimiento,
		JugadasPendientes: 	espaciosLibres(nuevoEstado, colorToca, defensa),
	}

	// Añadir a los hijos del padre
	nodo.Hijos = append(nodo.Hijos, nuevoNodo)
	return nuevoNodo
}

// Simulación------------------------------------------------------------------------------------------------------------------------------
func simularPartida(estadoBase *pb.GameState, miColor pb.PlayerColor, colorActual pb.PlayerColor, defensa bool, turnos int32) float64 {
	estado := clonarEstado(estadoBase)
	// Mientras no haya ganador o empate
	for true {
		ganador := obtenerGanador(estado)
		if ganador == miColor {
			return 1.0 // Gane
		}else if ganador != pb.PlayerColor_UNKNOWN{
			return 0.0 // Perdi
		}
		libres := espaciosLibres(estado, miColor, defensa)
		if len(libres) == 0 { // Revision de empate
			return 0.5 // Empate
		}

		jugada := libres[rand.Intn(len(libres))] // Elegir movimiento al azar
		
		jugar(estado, jugada, colorActual)
		
		if turnos % 2 == 1 {
			colorActual = colorOponente(colorActual) // Cambiamos de turno al bajar de nivel cuando queda 1 turno
			turnos++
		}else{ turnos--}
	}
	return 0.0
}

// Retropropagación -------------------------------------------------------------------------------------------------------------------------
func retropropagar(nodo *Nodo, resultado float64) {
	for nodo != nil {
		nodo.Visitas++
		nodo.Victorias += resultado
		
		// Si el nodo no está completado, revisamos si sus hijos ya lo están
		if !nodo.Completado && len(nodo.JugadasPendientes) == 0 && len(nodo.Hijos) > 0 {
			todosCompletados := true
			for _, hijo := range nodo.Hijos {
				if !hijo.Completado {
					todosCompletados = false // Encontramos un hijo que aún tiene ramas por explorar
					break
				}
			}
			nodo.Completado = todosCompletados // Si todos los hijos son true, el padre se vuelve true
		}
		nodo = nodo.Padre
	}
}

// Auxiliares--------------------------------------------------------------------------------------------------------------------------------
// Funcion que verifica los espacios en los que (inteligentemente) deberiamos jugar
func espaciosLibres(estado *pb.GameState, colorActual pb.PlayerColor, defensa bool) []*pb.Point {
	var libres []*pb.Point
	for x := int32(0); x < 19; x++ {
		for y := int32(0); y < 19; y++ {
			if estado.Board[x].Cells[y] == pb.PlayerColor_UNKNOWN {
				// Verifico modo defensa
				if defensa {
					if 	x_enLinea(x,y,estado,colorOponente(colorActual),3,[][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}) ||
						x_enLinea(x,y,estado,colorOponente(colorActual),3,[][2]int{{-1, 0}, {0, -1}, {-1, -1}, {-1, 1}}) {
						libres = append(libres, &pb.Point{X: x, Y: y})
					}	
				}else{
					if 	x_enLinea(x,y,estado,colorActual,1, [][2]int{{-1, 0}, {0, -1}, {-1, -1}, {-1, 1}}) ||
						x_enLinea(x,y,estado,colorActual,1, [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}})  {
						libres = append(libres, &pb.Point{X: x, Y: y})
					}	
				}
			}
		}
	}
	return libres
}

// Funcion para comparar tableros
func esMismoTablero(board1, board2 []*pb.Row) bool {
	for x := int32(0); x < 19; x++ {
		for y := int32(0); y < 19; y++ {
			if board1[x].Cells[y] != board2[x].Cells[y] {
				return false
			}
		}
	}
	return true
}

// Funcion que crea una copia exacta e independiente del estado del juego
func clonarEstado(estado *pb.GameState) *pb.GameState {
	// Copiamos los valores simples (números, booleanos, etc.)
	nuevo := &pb.GameState{
		Status:         estado.Status,
		MyColor:        estado.MyColor,
		StonesRequired: estado.StonesRequired,
	}

	// Copia del tablero
	if estado.Board != nil {
		nuevo.Board = make([]*pb.Row, len(estado.Board))
		for x, row := range estado.Board {
			nuevoRow := &pb.Row{
				Cells: make([]pb.PlayerColor, len(row.Cells)),
			}
			copy(nuevoRow.Cells, row.Cells)
			nuevo.Board[x] = nuevoRow
		}
	}
	return nuevo
}

// Funcion que coloca la piedra del color indicado en el tablero
func jugar(estado *pb.GameState, jugada *pb.Point, color pb.PlayerColor) {
	estado.Board[jugada.X].Cells[jugada.Y] = color
}

// Funcion que escanea el tablero buscando 5 piedras en línea, la piedra desde donde empezamos a contar + 5
func obtenerGanador(estado *pb.GameState) pb.PlayerColor {
	for x := int32(0); x < 19; x++ {
		for y := int32(0); y < 19; y++ {
			colorActual := estado.Board[x].Cells[y]
			if colorActual == pb.PlayerColor_UNKNOWN {
				continue // Celda vacía, pasamos a la siguiente
			}
			if x_enLinea(x,y,estado,colorActual,6,[][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}) {
				return colorActual
			}
		}
	}
	return pb.PlayerColor_UNKNOWN // Nadie ha ganado todavía
}

// Función para alternar colores de turnos
func colorOponente(color pb.PlayerColor) pb.PlayerColor {
	if color == pb.PlayerColor_BLACK {
		return pb.PlayerColor_WHITE
	}
	return pb.PlayerColor_BLACK
}

// Funcion que decide la jugada que vamos a hacer, mientras mas visitado es mas robusta la respuesta
func seleccionarHijoMasVisitado(nodo *Nodo) *Nodo {
	var mejorHijo *Nodo
	maxVisitas := -1						// Tengo que darle peso a lo mas relevante
	for _, hijo := range nodo.Hijos {
		if hijo.Visitas > maxVisitas {
			maxVisitas = hijo.Visitas
			mejorHijo = hijo
		}
	}
	return mejorHijo
}

// Funcion que verifica que haya limite en linea en las 4 direcciones: Derecha, Abajo, Diagonal-Abajo-Derecha, Diagonal-Arriba-Derecha
func x_enLinea(x int32, y int32, estado *pb.GameState, colorActual pb.PlayerColor, limite int, direccion [][2]int) bool{
	// Desde esta piedra en (x,y), miramos en las 4 direcciones
	for _, dir := range direccion {
		dx, dy := dir[0], dir[1]
		consecutivas := 1

		// Miramos hasta 4 posiciones más adelante
		for i := 1; i < 6; i++ {
			nx, ny := x+int32(dx*i), y+int32(dy*i)

			if nx < 0 || nx >= 19 || ny < 0 || ny >= 19 { // Si nos salimos del tablero, rompemos este bucle
				break
			}
			
			if estado.Board[nx].Cells[ny] == colorActual { // Si la piedra es del mismo color, sumamos
				consecutivas++
			}
		}
		// ¡Si llegamos a 6, este color acaba de ganar!
		if consecutivas >= limite {
			return true
		}
	}
	return false
}