package klog

import "testing"

func TestInfoTag(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	InfoTag(KlogTag("TestTAG"), "test")
	if !contains(infoLog, "I", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	if !contains(infoLog, "test", t) {
		t.Error("Info failed")
	}
}

func TestErrorTag(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	ErrorTag(KlogTag("TestTAG"), "test")
	if !contains(errorLog, "E", t) {
		t.Errorf("Error has wrong character: %q", contents(errorLog))
	}
	if !contains(errorLog, "test", t) {
		t.Error("Error failed")
	}
	str := contents(errorLog)
	if !contains(warningLog, str, t) {
		t.Error("Warning failed")
	}
	if !contains(infoLog, str, t) {
		t.Error("Info failed")
	}
}

func TestWarningTag(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	WarningTagf(KlogTag("TestTAG"), "test")
	if !contains(warningLog, "W", t) {
		t.Errorf("Warning has wrong character: %q", contents(warningLog))
	}
	if !contains(warningLog, "test", t) {
		t.Error("Warning failed")
	}
	str := contents(warningLog)
	if !contains(infoLog, str, t) {
		t.Error("Info failed")
	}
}
