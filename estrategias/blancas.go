package estrategias

import (
    pb "agentegabornelas/pb"
	"time"
)

// Creamos el bot a nivel de paquete para que NO pierda la memoria (el árbol) entre turnos.
var botBlancas = NuevoBot()

// Modo ataque
func JugarBlancas(estado *pb.GameState, turnos int32) []*pb.Point {

	if estado.Board[9].Cells[9] == pb.PlayerColor_UNKNOWN{	// Jugamos en el centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(9)},}
	}else if estado.Board[9].Cells[10] == pb.PlayerColor_UNKNOWN {	// Jugamos al lado del centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(10)},}
	}
    // Llamamos a MCTS a través de nuestro bot
    jugada := botBlancas.MCTS(estado, estado.MyColor, false, turnos)
	estado.Board[jugada[0].X].Cells[jugada[0].Y] = estado.MyColor
    
    return jugada
}