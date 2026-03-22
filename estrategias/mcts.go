package estrategias

import (
	"math"
	"time"
	"math/rand"
	pb "agentegabornelas/pb"
)

// Nodo representa un estado del tablero en nuestro árbol de búsqueda
type Nodo struct {
	Estado       		*pb.GameState 	// El estado del tablero en este nodo
	Padre       		*Nodo   	  	// El nodo padre
	Hijos     			[]*Nodo   		// Las posibles jugadas (estados) siguientes
	Jugada        		*pb.Point     	// La jugada que nos llevó a este estado
	Victorias      		float64      	// Cantidad de victorias simuladas desde aquí
	Visitas       		int           	// Cuántas veces hemos pasado por este nodo
	JugadasPendientes 	[]*pb.Point   	// Jugadas que aún no hemos explorado
	Completado 			bool			// Condicion de hijos explorados completamente
}

// Tree mantiene la memoria del árbol entre turnos
type Tree struct {
	Root *Nodo

}

// NuevoTree inicializa un arbol vacío
func NuevoTree() *Tree {
	return &Tree{Root: nil}
}

// MCTS actualiza el árbol y ejecuta MCTS---------------------------------------------------------------------------------------------------
func (Tree *Tree) MCTS(estadoActual *pb.GameState, miColor pb.PlayerColor, turnos int32) []*pb.Point {
	// Inicializar Raiz
	if Tree.Root == nil {
		Tree.Root = &Nodo{Estado: estadoActual, JugadasPendientes: espaciosLibres(estadoActual, miColor, turnos)}
	} else {
		// Buscamos si el estado actual ya fue evaluado y es un hijo de la raiz actual
		encontrado := false
		for _, hijo := range Tree.Root.Hijos {
			// Verificamos si el hijo de nuestra raiz es igual al estado actual
			if esMismoTablero(hijo.Estado.Board, estadoActual.Board) {
				Tree.Root = hijo
				Tree.Root.Padre = nil // Podamos el árbol por encima de la nueva raíz
				encontrado = true
				break
			}
		}
		// Si el rival hizo un movimiento no explorado
		if !encontrado {
			Tree.Root = &Nodo{Estado: estadoActual, JugadasPendientes: espaciosLibres(estadoActual, miColor, turnos)}
		}
	}

	// MCTS
	tiempo := time.Now().Add(4000*time.Millisecond)
	for time.Now().Before(tiempo) {

		if Tree.Root.Completado {
            break //Se completó el arbol, salimos del bucle antes
        }

		nodoActual := Tree.Root
		colorActual := miColor // Empezamos asumiendo que es nuestro turno en la Raíz
		// Selección
		for len(nodoActual.JugadasPendientes) == 0 && len(nodoActual.Hijos) > 0 {
			nodoActual = seleccion(nodoActual)
			if turnos % 2 == 1 {
				colorActual = ColorOponente(colorActual) // Cambiamos de turno al bajar de nivel cuando queda 1 turno
				turnos++
			}else{turnos--}
		}

		resultado := 0.0
		// Expansión
		if len(nodoActual.JugadasPendientes) > 0 {
			nodoActual = expandir(nodoActual, colorActual, turnos)
			if turnos % 2 == 1 {
				colorActual = ColorOponente(colorActual) // Cambiamos de turno al expandir cuando queda 1 turno
				turnos++
			}else{turnos--}
	
			// Simulación
			resultado = simularPartida(nodoActual.Estado, miColor, colorActual, turnos)
		}else{
			// No hay jugadas pendientes y no hay hijos. Es un estado terminal. Lo marcamos como completado.
            nodoActual.Completado = true
            
            // Como es un nodo terminal, 'simularPartida' solo verá quién ganó sin simular
            resultado = simularPartida(nodoActual.Estado, miColor, colorActual, turnos)
		}
		// Retropropagación
		retropropagar(nodoActual, resultado)
	}
	
	// Elegir el mejor movimiento (heuristica)
	mejorHijo := seleccionarHijoMasVisitado(Tree.Root)
	if mejorHijo == nil {
        return []*pb.Point{{X: 0, Y: 0}}// Si no hay hijos (tablero 100% lleno), mandamos una coordenada vacía de emergencia
    }
	// Actualizamos la raíz a la jugada que vamos a hacer
	Tree.Root = mejorHijo
	Tree.Root.Padre = nil
	return []*pb.Point{mejorHijo.Jugada}
}

// Selección--------------------------------------------------------------------------------------------------------------------------------
func seleccion(nodo *Nodo) *Nodo {
	// Aquí se aplica la fórmula matemática UCT
	// Balancea la "Explotación" (jugar lo que sabemos que es bueno) 
	// con la "Exploración" (probar movimientos nuevos).
	var mejorHijo *Nodo
	mejorPuntaje := math.Inf(-1)
	
	for _, hijo := range nodo.Hijos {
		// Fórmula UCB1 estándar: (Victorias / Visitas) + Raíz de 2 * sqrt(ln(VisitasPadre) / VisitasHijo)
		puntaje := (hijo.Victorias / float64(hijo.Visitas)) + math.Sqrt(2)*math.Sqrt(math.Log(float64(nodo.Visitas))/float64(hijo.Visitas))
		if puntaje > mejorPuntaje {
			mejorPuntaje = puntaje
			mejorHijo = hijo
		}
	}
	return mejorHijo
}

// Expansión--------------------------------------------------------------------------------------------------------------------------------
func expandir(nodo *Nodo, colorToca pb.PlayerColor, turnos int32) *Nodo {
	// Sacamos una jugada pendiente al azar
	indice := rand.Intn(len(nodo.JugadasPendientes))
	movimiento := nodo.JugadasPendientes[indice]

	// Eliminamos esa jugada de la lista (swap and pop)
	nodo.JugadasPendientes[indice] = nodo.JugadasPendientes[len(nodo.JugadasPendientes)-1]
	nodo.JugadasPendientes = nodo.JugadasPendientes[:len(nodo.JugadasPendientes)-1]

	// Clonamos el estado y aplico el movimiento con el color recibido
	nuevoEstado := clonarEstado(nodo.Estado)
	jugar(nuevoEstado, movimiento, colorToca)

	if turnos % 2 == 1 {
		colorToca = ColorOponente(colorToca) // Cambiamos de turno al expandir cuando queda 1 turno
	} 

	// Creamos el nuevo nodo
	nuevoNodo := &Nodo{
		Estado:        		nuevoEstado,
		Padre:       		nodo,
		Jugada:         	movimiento,
		JugadasPendientes: 	espaciosLibres(nuevoEstado, colorToca, turnos),
	}

	// Añadimos a los hijos del padre
	nodo.Hijos = append(nodo.Hijos, nuevoNodo)
	return nuevoNodo
}

// Simulación------------------------------------------------------------------------------------------------------------------------------
func simularPartida(estadoBase *pb.GameState, miColor, colorActual pb.PlayerColor, turnos int32) float64 {
	estado := clonarEstado(estadoBase)
	// Mientras no haya ganador o empate
	for true {
		ganador := obtenerGanador(estado)
		if ganador == miColor {
			return 1.0 // Gane
		}else if ganador != pb.PlayerColor_UNKNOWN{
			return 0.0 // Perdi
		}
		libres := espaciosLibres(estado, colorActual, turnos)
		if len(libres) == 0 { // Revision de empate
			return 0.5 // Empate
		}

		jugada := libres[rand.Intn(len(libres))] // Elegir movimiento al azar
		
		jugar(estado, jugada, colorActual)
		
		if turnos % 2 == 1 {
			colorActual = ColorOponente(colorActual) // Cambiamos de turno al bajar de nivel cuando queda 1 turno
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
func espaciosLibres(estado *pb.GameState, colorActual pb.PlayerColor, turnos int32) []*pb.Point {
	var espaciosVacios, mate, jaques, ataque1, ataque2, ataque3, ataque4, defensa, pre_jaque []*pb.Point
	//colorEnVentaja := colorActual
	//lineaMasLarga := 0
	for x := int32(0); x < 19; x++ {
		for y := int32(0); y < 19; y++ {
			if estado.Board[x].Cells[y] == ColorOponente(colorActual) {
				linead := Jaque(estado,ColorOponente(colorActual),3,x,y, [][2]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}})
				lineai := Jaque(estado,ColorOponente(colorActual),3,x,y, [][2]int{{-1, 0}, {0, -1}, {-1, -1}, {1, -1}})
				jaques = append(jaques, linead[0]...)
				jaques = append(jaques, lineai[0]...)

				pre_jaque = append(pre_jaque, linead[1]...)
				pre_jaque = append(pre_jaque, lineai[1]...)
			}else if estado.Board[x].Cells[y] == colorActual {
				// Ataque
				// Nos interesa aumentar nuestra linea mas larga
				linead := Jaque(estado,colorActual,4,x,y, [][2]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}})
				lineai := Jaque(estado,colorActual,4,x,y, [][2]int{{-1, 0}, {0, -1}, {-1, -1}, {1, -1}})
				mate = append(mate, linead[0]...)
				mate = append(mate, lineai[0]...)
				if len(mate) > 0 {
					return mate // Solo considera concretar jaque a favor
				}
				ataque4 = append(ataque4, linead[1]...)
				ataque4 = append(ataque4, lineai[1]...)
				if turnos > 1 && len(ataque4) > 0 {
					return ataque4 // Como jaque a favor porque queda otro turno
				}
				ataque3 = append(ataque3, linead[2]...)
				
				ataque3 = append(ataque3, lineai[2]...)
				ataque2 = append(ataque2, linead[3]...)
				ataque2 = append(ataque2, lineai[3]...)
			}else{
				// Ataque de 1
				derecha := X_enLinea(x,y,estado,colorActual,2,[][2]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}})
				izquierda := X_enLinea(x,y,estado,colorActual,2,[][2]int{{-1, 0}, {0, -1}, {-1, -1}, {1, -1}})
				if 	derecha[0] == 1 || derecha[1] == 1 || derecha[2] == 1 || derecha[3] == 1 ||
					izquierda[0] == 1 || izquierda[1] == 1 || izquierda[2] == 1 || izquierda[3] == 1 {
					ataque1 = append(ataque1, &pb.Point{X: x, Y: y})
				}	
				// Defensa
				// Nos interesa colocar fichas junto a las fichas del rival
				derecha = X_enLinea(x,y,estado,ColorOponente(colorActual),2,[][2]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}})
				izquierda = X_enLinea(x,y,estado,ColorOponente(colorActual),2,[][2]int{{-1, 0}, {0, -1}, {-1, -1}, {1, -1}})
				if 	derecha[0] == 1 || derecha[1] == 1 || derecha[2] == 1 || derecha[3] == 1 ||
					izquierda[0] == 1 || izquierda[1] == 1 || izquierda[2] == 1 || izquierda[3] == 1 {
					defensa = append(defensa, &pb.Point{X: x, Y: y})
				}
                espaciosVacios = append(espaciosVacios, &pb.Point{X: x, Y: y})
			}
		}
	}
	if len(jaques) > 0 {
		return jaques // Solo considera para trancar el jaque en contra
	}
	if len(pre_jaque) > 1 {
		return pre_jaque // Si hay mas de una linea de 3 del rival, eso quiere decir que mi rival puede hacer 2 lineas de 4 en la siguiente
	}
	ataque3 = filtrarAtaques(ataque4, ataque3)
	ataque2 = filtrarAtaques(ataque4, ataque2)
	ataque2 = filtrarAtaques(ataque3, ataque2)
	ataque1 = filtrarAtaques(ataque4, ataque1)
	ataque1 = filtrarAtaques(ataque3, ataque1)
	ataque1 = filtrarAtaques(ataque2, ataque1)
	if len(ataque3) > 0 {
        if !mismaLinea(ataque3) || len(ataque4) > 0 { // Tengo 2 lineas de 3, o 1 de 3 y otra de 4
            return ataque3
        }
	}
	if len(ataque2) > 0 {
		return ataque2
	}else if len(ataque1) > 0 {
		return ataque1
	}else if len(ataque4) > 0{
		return ataque4
	}else if len(defensa) > 0 {
		return defensa
	}
	return espaciosVacios //no deberia pasar
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

// Funcion que crea una copia del estado del juego
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
			linea := X_enLinea(x,y,estado,colorActual,6,[][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}})
			if linea[0] == 5 || linea[1] == 5 || linea[2] == 5 || linea[3] == 5 {
				return colorActual
			}
		}
	}
	return pb.PlayerColor_UNKNOWN // Nadie ha ganado todavía
}

// Función para alternar colores de turnos
func ColorOponente(color pb.PlayerColor) pb.PlayerColor {
	if color == pb.PlayerColor_BLACK {
		return pb.PlayerColor_WHITE
	}
	return pb.PlayerColor_BLACK
}

// Funcion que decide la jugada que vamos a hacer, mientras mas visitado es mas robusta la respuesta
func seleccionarHijoMasVisitado(nodo *Nodo) *Nodo {
	var mejorHijo *Nodo
	maxVisitas := -1
	for _, hijo := range nodo.Hijos {
		if hijo.Visitas > maxVisitas {
			maxVisitas = hijo.Visitas
			mejorHijo = hijo
		}
	}
	return mejorHijo
}

// Funcion que cuenta cuantas fichas hay en linea en las 4 direcciones en un "rango":
// 		Derecha, Abajo, Diagonal-Abajo-Derecha, Diagonal-Arriba-Derecha
//o		Izquierda, Arriba, Diagonal-Arriba-Izquierda, Diagonal-Abajo-Izquierda
func X_enLinea(x int32, y int32, estado *pb.GameState, colorActual pb.PlayerColor, rango int, direccion [][2]int) [4]int{
	direcciones := [4]int{}
	// Desde esta piedra en (x,y), revisamos en las 4 direcciones
	for c, dir := range direccion {
		dx, dy := dir[0], dir[1]
		consecutivas := 0
		// Revisamos las 5 posiciones siguientes
		for i := 1; i < rango; i++ {
			nx, ny := x+int32(dx*i), y+int32(dy*i)

			if nx < 0 || nx >= 19 || ny < 0 || ny >= 19 { // Si nos salimos del tablero, rompemos este bucle
				break
			}
			
			if estado.Board[nx].Cells[ny] == colorActual { // Si es del mismo color, sumamos
				consecutivas++
			}else if estado.Board[nx].Cells[ny] != pb.PlayerColor_UNKNOWN {
				break
			}
		}
		direcciones[c] = consecutivas
	}
	return direcciones
}


// Funcion que filtra el ataqueX quitandole los elementos del ataqueX+1
func filtrarAtaques(ataqueX1, ataqueX2 []*pb.Point) []*pb.Point {
	// Estructura para guardar los puntos
	type Punto struct {
		X, Y int32
	}
	mapa := make(map[Punto]struct{}, len(ataqueX1))
	for _, p := range ataqueX1 {
		if p != nil {
			key := Punto{X: p.X, Y: p.Y}
			mapa[key] = struct{}{}
		}
	}
	n := 0
	for _, p := range ataqueX2 {
		if p == nil {
			continue
		}
		key := Punto{X: p.X, Y: p.Y}
		// Si los valores de ataqueX2 no están en ataqueX1, lo conservamos
		if _, existe := mapa[key]; !existe {
			ataqueX2[n] = p
			n++
		}
	}
	// 3. Limpieza de punteros
	for i := n; i < len(ataqueX2); i++ {
		ataqueX2[i] = nil
	}
	return ataqueX2[:n]
}

// Funcion que verifica si todos los puntos del slice son colineales
func mismaLinea(puntos []*pb.Point) bool {
	if len(puntos) < 3 {
		return true
	}
	pA := puntos[0]
	pB := puntos[1]
	dx := int32(pB.X - pA.X)
	dy := int32(pB.Y - pA.Y)
	for i := 2; i < len(puntos); i++ {
		pC := puntos[i]
		if dy*int32(pC.X-pA.X) != dx*int32(pC.Y-pA.Y) {// Aplicamos el producto cruzado
			return false // Si la multiplicación no da igual, el punto NO está en la línea
		}
	}
	return true // Los puntos son colineales
}