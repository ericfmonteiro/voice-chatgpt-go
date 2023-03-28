package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"google.golang.org/api/option"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func main() {
	// Set up the context and client for the Speech-to-Text API.
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	client, err := speech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new audio stream using the microphone input.
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Set the audio configuration for the stream.
	req := &speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					Encoding:        speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz: 16000,
					LanguageCode:    "en-US",
				},
				InterimResults: true,
			},
		},
	}
	if err := stream.Send(req); err != nil {
		log.Fatalf("Could not send audio: %v", err)
	}

	// Start recording audio from the microphone.
	fmt.Println("Listening for audio...")
	audio := make([]byte, 1024)
	stop := make(chan bool)
	go func() {
		time.Sleep(3 * time.Second)
		stop <- true
	}()
	for {
		select {
		case <-stop:
			fmt.Println("Stopping recording...")
			return
		default:
			// Read audio from the microphone and send it to the Speech-to-Text API.
			if _, err := os.Stdin.Read(audio); err != nil {
				log.Fatalf("Could not read audio: %v", err)
			}
			req := &speechpb.StreamingRecognizeRequest{
				StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
					AudioContent: audio,
				},
			}
			if err := stream.Send(req); err != nil {
				log.Fatalf("Could not send audio: %v", err)
			}
		}
	}

	// Stop the audio stream and wait for the final result.
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("Could not close stream: %v", err)
	}
	resp, err := stream.Recv()
	if err != nil {
		log.Fatalf("Could not receive response: %v", err)
	}

	// Print the transcription result.
	for _, result := range resp.GetResults() {
		for _, alternative := range result.GetAlternatives() {
			fmt.Println(alternative.GetTranscript())
		}
	}
}
