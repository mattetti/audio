package generator

import "testing"

func TestOsc_Signal(t *testing.T) {
	osc := NewOsc(WaveSine, 440, 44100)
	if osc.CurrentPhaseAngle != 0 {
		t.Fatalf("expected the current phase to be zero")
	}
	if osc.phaseAngleIncr != 0.06268937721449021 {
		t.Fatalf("Wrong phase angle increment")
	}
	sample := osc.Sample()
	if phase := osc.CurrentPhaseAngle; phase != 0.06268937721449021 {
		t.Fatalf("wrong phase angle: %f, expected 0.06268937721449021", phase)
	}
	if sample != 0.0 {
		t.Fatalf("wrong first sample: %f expected 0.0", sample)
	}
	signal := osc.Signal(19)
	expected := []float32{0.06200187, 0.12406666, 0.18587168, 0.2471079, 0.30748, 0.36670643, 0.4245192, 0.4806642, 0.53490084, 0.58700234, 0.6367556, 0.6839611, 0.7284333, 0.77000004, 0.8085031, 0.8437978, 0.8757533, 0.90425223, 0.92919123}

	for i, s := range signal {
		if !nearlyEqual(s, expected[i], 0.000001) {
			t.Logf("sample %d didn't match, expected: %f got %f\n", i, expected[i], s)
			t.Fail()
		}
	}

	osc = NewOsc(WaveSine, 400, 1000)
	signal = osc.Signal(100)

	expected = []float32{0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599, 0, 0.5881599, -0.95136, 0.95136, -0.5881599}
	for i, s := range signal {
		if !nearlyEqual(s, expected[i], 0.00000001) {
			t.Logf("sample %d didn't match, expected: %f got %f\n", i, expected[i], s)
			t.Fail()
		}
	}
}
