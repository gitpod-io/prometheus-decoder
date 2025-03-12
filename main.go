package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

// PrometheusRecord is the ndjson-encoded format used for transporting metrics through firehose
type PrometheusRecord struct {
	Body []byte `json:"b"`
}

// TimeSeries represents a decoded Prometheus time series in a more readable format
type TimeSeries struct {
	Labels     map[string]string `json:"labels"`
	Timestamps []int64           `json:"timestamps"`
	Values     []float64         `json:"values"`
}

// WriteRequestJSON is a more readable representation of prompb.WriteRequest
type WriteRequestJSON struct {
	Timeseries []TimeSeries `json:"timeseries"`
}

// decodePrompbWriteReq decodes the wrapped prompb.WriteRequest
func decodePrompbWriteReq(record *PrometheusRecord) (*prompb.WriteRequest, error) {
	// Decompress the snappy-compressed data
	data, err := snappy.Decode(nil, record.Body)
	if err != nil {
		return nil, fmt.Errorf("snappy decode error: %w", err)
	}

	// Unmarshal the protobuf message
	var req prompb.WriteRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("protobuf unmarshal error: %w", err)
	}

	return &req, nil
}

// convertToReadableJSON converts a prompb.WriteRequest to a more readable JSON structure
func convertToReadableJSON(wreq *prompb.WriteRequest) *WriteRequestJSON {
	result := &WriteRequestJSON{
		Timeseries: make([]TimeSeries, len(wreq.Timeseries)),
	}

	for i, ts := range wreq.Timeseries {
		// Convert labels
		labels := make(map[string]string)
		for _, label := range ts.Labels {
			labels[label.Name] = label.Value
		}

		// Extract timestamps and values
		timestamps := make([]int64, len(ts.Samples))
		values := make([]float64, len(ts.Samples))
		for j, sample := range ts.Samples {
			timestamps[j] = sample.Timestamp
			values[j] = sample.Value
		}

		result.Timeseries[i] = TimeSeries{
			Labels:     labels,
			Timestamps: timestamps,
			Values:     values,
		}
	}

	return result
}

// humanReadableTime converts a Prometheus timestamp to a human-readable format
func humanReadableTime(timestamp int64) string {
	// Prometheus uses milliseconds since epoch
	return time.Unix(timestamp/1000, (timestamp%1000)*1000000).Format(time.RFC3339Nano)
}

// streamingJSONDecoder reads and processes a stream of JSON objects without requiring them to be newline-delimited
type streamingJSONDecoder struct {
	decoder *json.Decoder
	count   int
}

func newStreamingJSONDecoder(r io.Reader) *streamingJSONDecoder {
	decoder := json.NewDecoder(r)
	// Configure the decoder to support streams of concatenated JSON objects
	decoder.UseNumber()
	return &streamingJSONDecoder{
		decoder: decoder,
		count:   0,
	}
}

func (s *streamingJSONDecoder) next() (*PrometheusRecord, error) {
	var record PrometheusRecord
	if err := s.decoder.Decode(&record); err != nil {
		return nil, err
	}
	s.count++
	return &record, nil
}

func main() {
	inputFile := flag.String("input", "", "Input file containing PrometheusRecord entries (one per line)")
	outputFile := flag.String("output", "", "Output file for JSON results (default: stdout)")
	prettyPrint := flag.Bool("pretty", true, "Enable pretty-printing of JSON output")
	humanTime := flag.Bool("human-time", false, "Show human-readable timestamps in output")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: Input file is required")
		flag.Usage()
		os.Exit(1)
	}

	// Open input file
	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Prepare output
	var output *os.File
	if *outputFile == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(*outputFile)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer output.Close()
	}

	// Create JSON stream decoder
	decoder := newStreamingJSONDecoder(file)

	// Process each JSON object in the stream
	for {
		record, err := decoder.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error parsing JSON object #%d: %v\n", decoder.count, err)
			continue
		}

		// Decode the PrometheusRecord
		wreq, err := decodePrompbWriteReq(record)
		if err != nil {
			fmt.Printf("Error decoding JSON object #%d: %v\n", decoder.count, err)
			continue
		}

		// Convert to our more readable format
		jsonStruct := convertToReadableJSON(wreq)

		// Apply human-readable time conversion if requested
		if *humanTime {
			for i := range jsonStruct.Timeseries {
				humanTimes := make([]string, len(jsonStruct.Timeseries[i].Timestamps))
				for j, ts := range jsonStruct.Timeseries[i].Timestamps {
					humanTimes[j] = humanReadableTime(ts)
				}
				// We need to output this differently, so create a custom marshaling
				// This would require a custom struct and marshaling approach
				// For simplicity, we'll just add a note about it
				fmt.Fprintf(output, "# Object %d: Human-readable timestamps for reference:\n", decoder.count)
				for j, humanTime := range humanTimes {
					fmt.Fprintf(output, "#   Sample %d: %s\n", j, humanTime)
				}
			}
		}

		// Output the JSON
		var jsonData []byte
		if *prettyPrint {
			jsonData, err = json.MarshalIndent(jsonStruct, "", "  ")
		} else {
			jsonData, err = json.Marshal(jsonStruct)
		}

		if err != nil {
			fmt.Printf("Error encoding JSON object #%d to JSON: %v\n", decoder.count, err)
			continue
		}

		fmt.Fprintf(output, "# Object %d\n", decoder.count)
		fmt.Fprintln(output, string(jsonData))
	}

	fmt.Printf("Successfully processed %d JSON objects\n", decoder.count)
}
