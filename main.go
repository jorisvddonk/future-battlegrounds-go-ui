package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
	pb "github.com/jorisvddonk/future-battlegrounds-go-bot/futurebattlegrounds"
	v "github.com/ungerik/go3d/float64/vec2"
	"google.golang.org/grpc"
)

const (
	address      = "localhost:50051"
	screenWidth  = float64(1600)
	screenHeight = float64(800)
	zoomFactor   = float64(0.1)
)

func pv(v v.T) *v.T {
	return &v
}

func rv(v v.T) rl.Vector2 {
	return rl.NewVector2(float32(v[0]*zoomFactor+(screenWidth*0.5)), float32(v[1]*zoomFactor+(screenHeight*0.5)))
}

func main() {
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "Future Battlegrounds GO UI")

	rl.SetTargetFPS(60)

	var c pb.BattlegroundsClient
	var ships []*pb.Ship
	var bullets []*pb.Bullet

	// Set up a connection to the server.
	actualAddress := address
	if os.Getenv("FB_SERVER") != "" {
		actualAddress = os.Getenv("FB_SERVER")
	} else if len(os.Args) > 1 {
		arg := os.Args[1]
		actualAddress = arg
	}
	log.Printf("Connecting to %v", actualAddress)
	go func() {
		conn, err := grpc.Dial(actualAddress, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c = pb.NewBattlegroundsClient(conn)
		log.Printf("Connected!")

		// Stream battlegrounds
		ctx := context.Background()
		stream, err := c.StreamBattleground(ctx, &pb.Empty{})
		if err != nil {
			log.Fatalf("could not stream battleground: %v", err)
		}

		for {
			r, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("receiving battleground err: %v", err)
			}

			ships = r.Ships
			bullets = r.Bullets
		}
	}()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		text := fmt.Sprintf("Number of ships: %d", len(ships))

		for i := 0; i < len(ships); i++ {
			ship := ships[i]
			rot := v.T{ship.RotationVector.X, ship.RotationVector.Y}
			pos := v.T{ship.Position.X, ship.Position.Y}

			rl.DrawTriangleLines(
				rv(v.Add(pv(pos.Scaled(1)), pv(pv(rot.Scaled(1)).Rotate(0).Normalize().Scaled(100)))),
				rv(v.Add(pv(pos.Scaled(1)), pv(pv(rot.Scaled(1)).Rotate(math.Pi+0.5).Normalize().Scaled(100)))),
				rv(v.Add(pv(pos.Scaled(1)), pv(pv(rot.Scaled(1)).Rotate(math.Pi-0.5).Normalize().Scaled(100)))),
				rl.White)
		}

		for i := 0; i < len(bullets); i++ {
			bullet := bullets[i]
			pos := rv(v.T{bullet.Position.X, bullet.Position.Y})
			rl.DrawCircle(int32(pos.X), int32(pos.Y), 1, rl.Yellow)
		}

		rl.DrawText(text, 3, 3, 20, rl.LightGray)

		rl.EndDrawing()
	}

	rl.CloseWindow()

}
