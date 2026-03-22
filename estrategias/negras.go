package estrategias

import (
    pb "agentegabornelas/pb"
	"time"
	"log"
)

// Creamos el bot a nivel de paquete para que NO pierda la memoria (el árbol) entre turnos.
var botNegras = NuevoBot()

// Modo defensa
func JugarNegras(estado *pb.GameState, turnos int32 ) []*pb.Point {
    
	if estado.Board[9].Cells[9] == pb.PlayerColor_UNKNOWN{	// Jugamos en el centro
		time.Sleep(500*time.Millisecond)
		return []*pb.Point{&pb.Point{X: int32(9), Y: int32(9)},}
	}
	
	//Llamamos a MCTS a través de nuestro bot
	jugada := botNegras.MCTS(estado, estado.MyColor, true, turnos)
	log.Printf("El Movimiento es: %v", jugada)
	return jugada
}
// ALGO CON EL JAQUE
// Funcion que verifica si el estado es de jaque, es decir, si el rival tiene 4 o 5 en linea
// devuelve una posicion x,y para jugar o nil si no hay jaque
func Jaque(estado *pb.GameState, colorRival pb.PlayerColor, x, y int32) []*pb.Point{
	// Busco en el tablero, si X_enLinea[i] da true, ese x,y esta en jaque en la linea i
	var trancar []*pb.Point
	linea := X_enLinea(x,y,estado,colorRival,3,6,[][2]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}})
	// Casos
	if linea[0]{ // Abajo
		trancar = tieneEspacioPara6(estado, colorRival,x+1,y,x+2,y)// revisa en medio
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x-1,y,x-2,y)// revisa antes
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x+4,y,x+5,y)// revisa despues
		if len(trancar) > 0 {return trancar}
	}else if linea[1]{ // derecha
		trancar = tieneEspacioPara6(estado, colorRival,x,y+1,x,y+2)// revisa en medio
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x,y-1,x,y-2)// revisa antes
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x,y+4,x,y+5)// revisa despues
		if len(trancar) > 0 {return trancar}
	}else if linea[2]{ // abajo-derecha
		trancar = tieneEspacioPara6(estado, colorRival,x+1,y+1,x+2,y+2)// revisa en medio
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x-1,y-1,x-2,y-2)// revisa antes
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x+4,y+4,x+5,y+5)// revisa despues
		if len(trancar) > 0 {return trancar}
	}else if linea[3]{ // arriba-derecha
		trancar = tieneEspacioPara6(estado, colorRival,x-1,y+1,x-2,y+2)// revisa en medio
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x+1,y-1,x+2,y-2)// revisa antes
		if len(trancar) > 0 {return trancar}
		trancar = tieneEspacioPara6(estado, colorRival,x-4,y+4,x-5,y+5)// revisa despues
		if len(trancar) > 0 {return trancar}
	}
	return nil
}

// Funcion que verifica si el rival puede llegar a 6 fichas consecutivas a un lado de las 4 o 5 que ya tiene o en medio
// devuelve la posicion x,y a jugar para trancar o nil si ya fue trancaso (no hay jaque)
func tieneEspacioPara6(estado *pb.GameState, colorRival pb.PlayerColor, x, y, x2, y2 int32) []*pb.Point{
	if x >= 0 && x < 19 && y >= 0 && y < 19 && x2 >= 0 && x2 < 19 && y2 >= 0 && y2 < 19 {
		if estado.Board[x].Cells[y] == pb.PlayerColor_UNKNOWN {
			return []*pb.Point{&pb.Point{X: x, Y: y}}
		}else if estado.Board[x].Cells[y] == ColorOponente(colorRival){
			return nil // si ya tranque, vuelve
		}else if estado.Board[x2].Cells[y2] == pb.PlayerColor_UNKNOWN {
			return []*pb.Point{&pb.Point{X: x2, Y: y2}}
		}
	}
	return nil
}

// quiero que mi ficha este en linea con otra mia pero esa linea debe ser cruzada con la linea del rival
// si eso no se consigue, que lo ponga junto a uno del rival
func AtaqueTimido(estado *pb.GameState, miColor pb.PlayerColor, x, y int32) bool{
	miLineaD := X_enLinea(x,y,estado,miColor,1,3,[][2]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}})
	miLineaI := X_enLinea(x,y,estado,miColor,1,3,[][2]int{{-1, 0}, {0, -1}, {-1, -1}, {1, -1}})
	miLineantD:= X_enLinea(x,y,estado,ColorOponente(miColor),1,3,[][2]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}})
	miLineantI:= X_enLinea(x,y,estado,ColorOponente(miColor),1,3,[][2]int{{-1, 0}, {0, -1}, {-1, -1}, {1, -1}})
	if 	( (miLineaD[0] || miLineaI[0]) && (miLineantD[1] || miLineantD[2] || miLineantD[3] || miLineantI[1] || miLineantI[2] || miLineantI[3]) ) || 
		( (miLineaD[1] || miLineaI[1]) && (miLineantD[0] || miLineantD[2] || miLineantD[3] || miLineantI[0] || miLineantI[2] || miLineantI[3]) ) ||
		( (miLineaD[2] || miLineaI[2]) && (miLineantD[1] || miLineantD[0] || miLineantD[3] || miLineantI[1] || miLineantI[0] || miLineantI[3]) ) ||
		( (miLineaD[3] || miLineaI[3]) && (miLineantD[1] || miLineantD[2] || miLineantD[0] || miLineantI[1] || miLineantI[2] || miLineantI[0]) ) {
		return true
	}
	return false
}