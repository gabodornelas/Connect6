package estrategias

import (
    pb "agentegabornelas/pb"
	"time"
)

// Creamos el arbol a nivel de paquete para que NO pierda la memoria entre turnos.
var tree = NuevoTree()

// Modo defensa
func Jugar(estado *pb.GameState, turnos int32) []*pb.Point {
    
	if estado.Board[9].Cells[9] == pb.PlayerColor_UNKNOWN{	// Jugamos en el centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(9)},}
	}else if estado.Board[9].Cells[10] == pb.PlayerColor_UNKNOWN {	// Jugamos al lado del centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(10)},}
	}
	
	//Llamamos a MCTS a través de nuestro bot
	jugada := tree.MCTS(estado, estado.MyColor, turnos)
	return jugada
}



// Funcion que verifica si el estado es de jaque, es decir, si el rival tiene 4 o 5 en linea
// 		para 3 o menos en linea tambien es util
// devuelve una posicion x,y para jugar o nil si no hay jaque
func Jaque(estado *pb.GameState, colorRival pb.PlayerColor, alineadas int, x, y int32, direcciones [][2]int) [4][]*pb.Point {
	// Busco en el tablero, si X_enLinea[i] da alineadas-limite, ese x,y esta en el jaque que nos interesa en la direccion i
	var trancar [4][]*pb.Point
	linea := X_enLinea(x, y, estado, colorRival, 6, direcciones)
	for limite := 0; limite < alineadas; limite++ {
		for dirIndex, dir := range direcciones {
			if linea[dirIndex] == alineadas-limite {
				dx, dy := int32(dir[0]), int32(dir[1])
				// Posiciones relativas dinámicamente usando dx y dy
				intento := [10]int32{
					x+dx, y+dy, x+(dx*2), y+(dy*2), x+(dx*3), y+(dy*3),	// Revisa en medio
					x+(dx*4), y+(dy*4), x+(dx*5), y+(dy*5),				// Revisa adelante
				}
				trancar[limite] = tieneEspacioPara6(estado, colorRival, intento[0], intento[1], intento[2], intento[3],		// +1, +2
																		intento[4], intento[5], intento[6], intento[7],		// +3, +4
																		intento[8], intento[9])								// +5
				if len(trancar[limite]) > 0 {
					break // ya trancamos, no interesan las otras direcciones
				}
			}
		}
	}
	return trancar
}

// Funcion que verifica si el rival puede llegar a 6 fichas consecutivas a un lado de las 4 o 5 que ya tiene o en medio
// devuelve la posicion x,y a jugar para trancar o nil si ya fue trancaso (no hay jaque)
func tieneEspacioPara6(estado *pb.GameState, colorRival pb.PlayerColor, x1, y1, x2, y2, x3, y3, x4, y4, x5, y5 int32) []*pb.Point{
	if 	x1 >= 0 && x1 < 19 && y1 >= 0 && y1 < 19 && x2 >= 0 && x2 < 19 && y2 >= 0 && y2 < 19 && x3 >= 0 && x3 < 19 && y3 >= 0 && y3 < 19 &&
		x4 >= 0 && x4 < 19 && y4 >= 0 && y4 < 19 && x5 >= 0 && x5 < 19 && y5 >= 0 && y5 < 19 &&
		estado.Board[x1].Cells[y1] != ColorOponente(colorRival) && estado.Board[x2].Cells[y2] != ColorOponente(colorRival) &&
		estado.Board[x3].Cells[y3] != ColorOponente(colorRival) && estado.Board[x4].Cells[y4] != ColorOponente(colorRival) &&
		estado.Board[x5].Cells[y5] != ColorOponente(colorRival){
		if estado.Board[x1].Cells[y1] == pb.PlayerColor_UNKNOWN {
			return []*pb.Point{&pb.Point{X: x1, Y: y1}}
		}else if estado.Board[x2].Cells[y2] == pb.PlayerColor_UNKNOWN {
			return []*pb.Point{&pb.Point{X: x2, Y: y2}}
		}else if estado.Board[x3].Cells[y3] == pb.PlayerColor_UNKNOWN {
			return []*pb.Point{&pb.Point{X: x3, Y: y3}}
		}else if estado.Board[x4].Cells[y4] == pb.PlayerColor_UNKNOWN {
			return []*pb.Point{&pb.Point{X: x4, Y: y4}}
		}else if estado.Board[x5].Cells[y5] == pb.PlayerColor_UNKNOWN {
			return []*pb.Point{&pb.Point{X: x5, Y: y5}}
		}
	}
	return nil // ya fue trancado
}