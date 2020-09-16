package klog

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-logr/logr"
)

type KlogTag string

// InfoTag logs to the INFO log with tag.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func InfoTag(tag KlogTag, args ...interface{}) {
	logging.printTag(tag, infoLog, logging.logr, args...)
}

func InfoTagDepth(tag KlogTag, depth int, args ...interface{}) {
	logging.printDepthTag(tag, infoLog, logging.logr, depth, args...)
}

func InfoTagln(tag KlogTag, args ...interface{}) {
	logging.printTagln(tag, infoLog, logging.logr, args...)
}

func InfoTagf(tag KlogTag, format string, args ...interface{}) {
	logging.printTagf(tag, infoLog, logging.logr, format, args...)
}

func WarningTag(tag KlogTag, args ...interface{}) {
	logging.printTag(tag, warningLog, logging.logr, args...)
}

func WarningTagln(tag KlogTag, args ...interface{}) {
	logging.printTagln(tag, warningLog, logging.logr, args...)
}

func WarningTagf(tag KlogTag, format string, args ...interface{}) {
	logging.printTagf(tag, warningLog, logging.logr, format, args...)
}

func ErrorTag(tag KlogTag, args ...interface{}) {
	logging.printTag(tag, errorLog, logging.logr, args...)
}

func ErrorTagln(tag KlogTag, args ...interface{}) {
	logging.printTagln(tag, errorLog, logging.logr, args...)
}

func ErrorTagf(tag KlogTag, format string, args ...interface{}) {
	logging.printTagf(tag, errorLog, logging.logr, format, args...)
}

func (l *loggingT) printTag(tag KlogTag, s severity, logr logr.Logger, args ...interface{}) {
	l.printDepthTag(tag, s, logr, 1, args...)
}

func (l *loggingT) printTagf(tag KlogTag, s severity, logr logr.Logger, format string, args ...interface{}) {
	buf, file, line := l.headerTag(tag, s, 0)
	// if logr is set, we clear the generated header as we rely on the backing
	// logr implementation to print headers
	if logr != nil {
		l.putBuffer(buf)
		buf = l.getBuffer()
	}
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, logr, buf, file, line, false)
}

func (l *loggingT) printTagln(tag KlogTag, s severity, logr logr.Logger, args ...interface{}) {
	buf, file, line := l.headerTag(tag, s, 0)
	// if logr is set, we clear the generated header as we rely on the backing
	// logr implementation to print headers
	if logr != nil {
		l.putBuffer(buf)
		buf = l.getBuffer()
	}
	fmt.Fprintln(buf, args...)
	l.output(s, logr, buf, file, line, false)
}

func (l *loggingT) printDepthTag(tag KlogTag, s severity, logr logr.Logger, depth int, args ...interface{}) {
	buf, file, line := l.headerTag(tag, s, depth)
	// if logr is set, we clear the generated header as we rely on the backing
	// logr implementation to print headers
	if logr != nil {
		l.putBuffer(buf)
		buf = l.getBuffer()
	}
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, logr, buf, file, line, false)
}

/*
headerTag formats a log header with tag as defined by the C++ implementation.
It returns a buffer containing the formatted header and the user's file and line number.
The depth specifies how many stack frames above lives the source line to be identified in the log message.

Log lines have this form:
	Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...
where the fields are defined as follows:
	L                A single character, representing the log level (eg 'I' for INFO)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	file             The file name
	line             The line number
	tag              The tag
	msg              The user-supplied message
*/
func (l *loggingT) headerTag(tag KlogTag, s severity, depth int) (*buffer, string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		if slash := strings.LastIndex(file, "/"); slash >= 0 {
			path := file
			file = path[slash+1:]
			if l.addDirHeader {
				if dirsep := strings.LastIndex(path[:slash], "/"); dirsep >= 0 {
					file = path[dirsep+1:]
				}
			}
		}
	}
	return l.formatHeaderTag(tag, s, file, line), file, line
}

// formatHeaderTag formats a log header using the provided file name and line number and tag.
func (l *loggingT) formatHeaderTag(tag KlogTag, s severity, file string, line int) *buffer {
	now := timeNow()
	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}
	if s > fatalLog {
		s = infoLog // for safety.
	}
	buf := l.getBuffer()
	if l.skipHeaders {
		return buf
	}

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	// year, month, day := now.Date()
	// hour, minute, second := now.Clock()
	// Lmmdd hh:mm:ss.uuuuuu threadid file:line]

	tmp := now.Format("2006-01-02 15:04:05.999 MST")
	buf.WriteString(tmp)
	buf.tmp[0] = ' '
	buf.nDigits(7, 1, pid, ' ') // TODO: should be TID
	buf.tmp[8] = ' '
	buf.tmp[9] = '['
	buf.Write(buf.tmp[:10])
	buf.WriteString(severityName[s])
	buf.tmp[0] = ']'
	buf.tmp[1] = ' '
	buf.Write(buf.tmp[:2])
	buf.WriteString(file)
	buf.tmp[0] = ':'
	n := buf.someDigits(1, line)
	buf.tmp[n+1] = ']'
	buf.tmp[n+2] = ' '
	buf.Write(buf.tmp[:n+3])
	buf.WriteString(string(tag))
	buf.tmp[0] = ':'
	buf.tmp[1] = ' '
	buf.Write(buf.tmp[:2])
	return buf
}
