package estrategias

import (
    pb "agentegabornelas/pb"
	"time"
)

// 1. Creamos el bot a nivel de paquete para que NO pierda la memoria (el árbol) entre turnos.
var botBlancas = NuevoBot()

// Modo ataque
func JugarBlancas(state *pb.GameState, turnos int32) []*pb.Point {

	if state.Board[9].Cells[9] == pb.PlayerColor_UNKNOWN{	// Jugamos en el centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(9)},}
	}else if state.Board[9].Cells[10] == pb.PlayerColor_UNKNOWN {	// Jugamos al lado del centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(10)},}
	}
    // Llamamos a MCTS a través de nuestro bot
    jugada := botBlancas.MCTS(state, state.MyColor, false, turnos)
    
    return jugada
}