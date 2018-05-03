package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/republicprotocol/republic-go/rpc/status"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func main() {
	// Create new cli application
	app := cli.NewApp()

	// Define sub-commands
	app.Commands = []cli.Command{
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "get status of the node with given address",
			Action: func(c *cli.Context) error {
				return GetStatus(c.Args())
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func GetStatus(args []string) error {
	// Check arguments length
	if len(args) != 1 {
		return fmt.Errorf("please provide target network address (e.g. 0.0.0.0:8080)")
	}
	address := args[0]

	// Dial to the target node
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c := status.NewStatusClient(conn)

	// Call status rpc
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rsp, err := c.Status(ctx, &status.StatusRequest{})
	if err != nil {
		return err
	}

	// Output the result as a json string
	output, err := json.Marshal(rsp)
	if err != nil {
		return err
	}
	log.Println(string(output))

	return nil
}
