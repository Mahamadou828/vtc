// Package sentryfmt is a wrapper around the sentry package. This package format
// and send the errors to the sentry hub. The events are formatted to included
// more data and useful information
package lambda

import (
	"github.com/getsentry/sentry-go"
	"reflect"
	"strings"
)

func CaptureError(rt *RequestTrace, statusCode int, err error) {
	event := eventFromException(err, sentry.LevelError)
	//send error data
	event.Exception[0].Type = rt.Path
	event.Message = err.Error()
	//attach context information
	event.Contexts["request.trace"] = sentry.Context{
		"id":         rt.ID,
		"now":        rt.Now,
		"statusCode": statusCode,
	}

	//capturing the event
	sentry.CaptureEvent(event)
}

func eventFromException(exception error, level sentry.Level) *sentry.Event {
	err := exception

	event := sentry.NewEvent()
	event.Level = level

	for i := 0; i < 10 && err != nil; i++ {
		event.Exception = append(event.Exception, sentry.Exception{
			Value:      err.Error(),
			Type:       reflect.TypeOf(err).String(),
			Stacktrace: sentry.ExtractStacktrace(err),
		})
		switch previous := err.(type) {
		case interface{ Unwrap() error }:
			err = previous.Unwrap()
		case interface{ Cause() error }:
			err = previous.Cause()
		default:
			err = nil
		}
	}

	// Add a trace of the current stack to the most recent error in a chain if
	// it doesn't have a stack trace yet.
	// We only add to the most recent error to avoid duplication and because the
	// current stack is most likely unrelated to errors deeper in the chain.
	if event.Exception[0].Stacktrace == nil {
		st := sentry.NewStacktrace()
		filteredFrames := make([]sentry.Frame, len(st.Frames))
		for _, f := range st.Frames {
			//ignore vendor frames
			if strings.Contains(f.AbsPath, "vendor") {
				continue
			}
			//ignore native go package
			if strings.Contains(f.AbsPath, "/Go/src") {
				continue
			}
			if f.Function == "eventFromException" {
				continue
			}
			filteredFrames = append(filteredFrames, f)
		}
		st.Frames = filteredFrames
		event.Exception[0].Stacktrace = st
	}

	// event.Exception should be sorted such that the most recent error is last.
	reverse(event.Exception)

	return event
}

// reverse reverses the slice a in place.
func reverse(a []sentry.Exception) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
