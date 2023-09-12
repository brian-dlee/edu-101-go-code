package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/alecthomas/kingpin/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	farewell "temporal101/exercises/farewell-workflow/solution"
)

func main() {
	var (
		cli = kingpin.New("replay-101-farewell-workflow", "")

		// execute worker
		cmdWorker = cli.Command("worker", "Register a new user.")

		// execute-greet-someone
		cmdXGS        = cli.Command("xgs", "Execute greet someone.")
		cmdXGSArgName = cmdXGS.Arg("name", "The name of the person to greeting").String()
	)

	switch kingpin.MustParse(cli.Parse(os.Args[1:])) {
	case cmdWorker.FullCommand():
		runWorker()
	case cmdXGS.FullCommand():
		executeWorker(*cmdXGSArgName)
	}
}

func createWorker(c client.Client) worker.Worker {
	w := worker.New(c, "greeting-tasks", worker.Options{})

	w.RegisterWorkflow(farewell.GreetSomeone)
	w.RegisterActivity(farewell.GreetInSpanish)
	w.RegisterActivity(farewell.FarewellInSpanish)

	return w
}

func runWorker() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Panicf("failed to create client connection: %s", err)
	}
	defer c.Close()

	w := createWorker(c)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func executeWorker(name string) {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Panicf("failed to create client connection: %s", err)
	}
	defer c.Close()

	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("greet-%s", slugify(name)),
		TaskQueue: "greeting-tasks",
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, farewell.GreetSomeone, name)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}

	log.Println("Workflow result:", result)
}

func slugify(text string) string {
	re := regexp.MustCompile("[^-a-zA-Z0-9]+")
	return re.ReplaceAllString(text, "-")
}
