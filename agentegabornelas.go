package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	pb "agentegabornelas/pb"

	"agentegabornelas/estrategias"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func playGame(addr, teamName string) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Error al conectar: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewGameServerClient(conn)
	stream, err := client.Play(context.Background())
	if err != nil {
		log.Printf("Error al abrir stream: %v", err)
		return
	}

	// Registro del equipo
	log.Printf("Registrando equipo: %s", teamName)
	stream.Send(&pb.PlayerAction{
		Action: &pb.PlayerAction_RegisterTeam{RegisterTeam: teamName},
	})

	// Bucle principal de juego
	for {
		state, err := stream.Recv()
		if err != nil {
			log.Println("🔌 Conexión cerrada por el servidor")
			return
		}

		switch state.Status {
		case pb.GameState_WAITING:
			log.Println("⏳ Esperando contrincante...")

		case pb.GameState_PLAYING:
			if state.IsMyTurn {
				log.Printf("🎲 Es mi turno (%v). Piedras requeridas: %d", state.MyColor, state.StonesRequired)

				stones := []*pb.Point{}
				
				for int32(len(stones)) < state.StonesRequired {
					stones = append(stones, estrategias.Jugar(state, state.StonesRequired - int32(len(stones)))...)
					if int32(len(stones)) < 2{
						state.Board[stones[0].X].Cells[stones[0].Y] = state.MyColor
					}else{
						state.Board[stones[1].X].Cells[stones[1].Y] = state.MyColor
					}
				}
				// Enviar movimiento
				err := stream.Send(&pb.PlayerAction{
					Action: &pb.PlayerAction_Move{
						Move: &pb.Move{Stones: stones},
					},
				})
				if err != nil {
					log.Printf("Error al enviar movimiento: %v", err)
					return
				}
				log.Printf("✅ Movimiento enviado: %v", stones)
			} else {
				log.Println("⌛ Esperando a que el oponente mueva...")
			}

		case pb.GameState_FINISHED:
			log.Printf("🏁 PARTIDA FINALIZADA")
			log.Printf("🏆 Ganador: %v", state.Winner)
			log.Printf("📝 Resultado: %v", state.Result)
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "servidor:50051"
	}

	teamName := os.Getenv("TEAM_NAME")
	if teamName == "" {
		teamName = "GabrielBot"
	}

	// Loop de reconexión: después de cada partida, reconectar
	for {
		log.Printf("🔄 Conectando a %s como %s...", addr, teamName)
		playGame(addr, teamName)
		log.Println("⏳ Reconectando en 3 segundos...")
		time.Sleep(3 * time.Second)
	}
}