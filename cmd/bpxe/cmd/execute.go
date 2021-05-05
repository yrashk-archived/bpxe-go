package cmd

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	"github.com/spf13/cobra"
)

// executeCmd represents the execute command
var executeCmd = &cobra.Command{
	Use:   "execute [file.bpmn]",
	Short: "Execute BPMN model",
	Long:  `This command will execute processes in a BPMN model.`,
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		var document bpmn.Definitions
		var err error
		src, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Can't read file: %v\n", err)
			return
		}
		err = xml.Unmarshal(src, &document)
		if err != nil {
			fmt.Printf("XML unmarshalling error: %v\n", err)
			return
		}
		for i := range *document.Processes() {
			processElement := &(*document.Processes())[i]
			if id, present := processElement.Id(); present {
				fmt.Printf("Loaded process %s\n", *id)
			} else {
				fmt.Println("Loaded an unnamed process")
			}
			proc := process.NewProcess(processElement, &document)
			if instance, err := proc.Instantiate(); err == nil {
				traces := instance.Tracer.Subscribe()
				err := instance.Run()
				if err != nil {
					fmt.Printf("failed to run the instance: %s\n", err)
				}
				done := make(chan bool)
				go func() {
					for {
						trace := <-traces
						switch trace := trace.(type) {
						case flow.NewFlowTrace:
							fmt.Printf("New flow %s\n", trace.FlowId.String())
						case flow.FlowTrace:
							sourceId, present := trace.Source.Id()
							if !present {
								sourceId = new(string)
								*sourceId = "unnamed"
							}
							for _, flow := range trace.Flows {
								target, err := flow.SequenceFlow().Target()
								if err != nil {
									fmt.Printf("Can't find target in a flow")
								}
								targetId, present := target.Id()
								if !present {
									targetId = new(string)
									*targetId = "unnamed"
								}
								fmt.Printf("Flow(%s) %s -> %s\n", flow.Id().String(), *sourceId, *targetId)
							}
						case flow.CeaseFlowTrace:
							fmt.Printf("No flows left\n")
							done <- true
							return
						case tracing.ErrorTrace:
							fmt.Printf("Error: %v\n", trace.Error)
						default:
						}

					}
				}()
				instance.WaitUntilComplete(context.Background())
				<-done
			} else {
				fmt.Printf("failed to instantiate the process: %s\n", err)
			}
		}

	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(executeCmd)
}
