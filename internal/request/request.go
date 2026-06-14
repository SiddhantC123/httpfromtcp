package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"httpserver/internal/headers"
)

const (
	stateInitialized = iota
	stateParsingHeaders
	stateParsingBody
	stateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	const bufferSize = 1024
	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{
		state:   stateInitialized,
		Headers: headers.NewHeaders(),
	}

	for req.state != stateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, fmt.Errorf("failed to read request: %w", err)
		}

		readToIndex += n

		consumed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if consumed > 0 {
			copy(buf, buf[consumed:readToIndex])
			readToIndex -= consumed
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != stateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case stateInitialized:
		reqLine, consumed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}

		r.RequestLine = reqLine
		r.state = stateParsingHeaders
		return consumed, nil

	case stateParsingHeaders:
		consumed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			// STEP 4: Transition to parsing the body instead of done!
			r.state = stateParsingBody
		}
		return consumed, nil

	case stateParsingBody:

		lengthStr := r.Headers.Get("Content-Length")
		if lengthStr == "" {

			r.state = stateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length: %w", err)
		}

		bytesNeeded := contentLength - len(r.Body)

		// 4. Figure out how many bytes we can ACTUALLY read from the current data chunk
		bytesToRead := bytesNeeded
		if len(data) < bytesNeeded {
			bytesToRead = len(data)
		}

		// 5. Append what we have to the body slice
		r.Body = append(r.Body, data[:bytesToRead]...)

		// 6. If we've collected the full payload, the request is fully parsed
		if len(r.Body) == contentLength {
			r.state = stateDone
		}

		// Consume the bytes we just appended
		return bytesToRead, nil

	case stateDone:
		return 0, errors.New("error: trying to read data in a done state")

	default:
		return 0, errors.New("error: unknown state")
	}
}

func parseRequestLine(data string) (RequestLine, int, error) {
	idx := strings.Index(data, "\r\n")
	if idx == -1 {
		return RequestLine{}, 0, nil
	}

	line := data[:idx]
	consumed := idx + 2

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line: expected 3 parts, got %d", len(parts))
	}

	method := parts[0]
	target := parts[1]
	versionRaw := parts[2]

	for _, char := range method {
		if char < 'A' || char > 'Z' {
			return RequestLine{}, 0, fmt.Errorf("invalid method: %s contains non-uppercase characters", method)
		}
	}

	if versionRaw != "HTTP/1.1" {
		return RequestLine{}, 0, fmt.Errorf("unsupported HTTP version: %s", versionRaw)
	}

	return RequestLine{Method: method, RequestTarget: target, HttpVersion: "1.1"}, consumed, nil
}
