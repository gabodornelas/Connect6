package estrategias

import (
    pb "agentegabornelas/pb"
	"time"
)

// 1. Creamos el bot a nivel de paquete para que NO pierda la memoria (el árbol) entre turnos.
var botNegras = NuevoBot()

// Modo defensa
func JugarNegras(state *pb.GameState, turnos int32 ) []*pb.Point {
    
	if state.Board[9].Cells[9] == pb.PlayerColor_UNKNOWN{	// Jugamos en el centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(9)},}
	}
	//Printf("nodo raiz es el que vino de : %v", botNegras.Root.Jugada)
    // Llamamos a MCTS a través de nuestro bot
    jugada := botNegras.MCTS(state, state.MyColor, true, turnos)
    
    return jugada
}