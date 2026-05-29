package uiprefs

import "testing"

func TestReadLanguageDefaultsToEnglish(t *testing.T) {
	got := ReadLanguage("unknown")
	if got != LangEN {
		t.Fatalf("expected %q, got %q", LangEN, got)
	}
}

func TestReadLanguageAcceptsSupportedValue(t *testing.T) {
	got := ReadLanguage("es")
	if got != LangES {
		t.Fatalf("expected %q, got %q", LangES, got)
	}
}

func TestReadThemeDefaultsToSystem(t *testing.T) {
	got := ReadTheme("whatever")
	if got != "System" {
		t.Fatalf("expected System, got %q", got)
	}
}

func TestLanguageOptionsContainsAllExpectedEntries(t *testing.T) {
	opts := LanguageOptions()
	if len(opts) != 5 {
		t.Fatalf("expected 5 language options, got %d", len(opts))
	}
}

// ---------- font & density ---------------------------------------------------

func TestReadFontSizeDefaultsWhenEmpty(t *testing.T) {
	got := ReadFontSize("")
	if got != DefaultFontSize {
		t.Fatalf("expected %v, got %v", DefaultFontSize, got)
	}
}

func TestReadFontSizeDefaultsOnInvalidValue(t *testing.T) {
	got := ReadFontSize("abc")
	if got != DefaultFontSize {
		t.Fatalf("expected %v, got %v", DefaultFontSize, got)
	}
}

func TestReadFontSizeDefaultsWhenOutOfRange(t *testing.T) {
	got := ReadFontSize("99")
	if got != DefaultFontSize {
		t.Fatalf("expected %v, got %v", DefaultFontSize, got)
	}
}

func TestReadFontSizeAcceptsValidValue(t *testing.T) {
	got := ReadFontSize("18")
	if got != 18 {
		t.Fatalf("expected 18, got %v", got)
	}
}

func TestReadDensityDefaultsToNormal(t *testing.T) {
	got := ReadDensity("unknown")
	if got != DefaultDensity {
		t.Fatalf("expected %q, got %q", DefaultDensity, got)
	}
}

func TestReadDensityAcceptsKnownValues(t *testing.T) {
	for _, v := range DensityOptions() {
		got := ReadDensity(v)
		if got != v {
			t.Fatalf("expected %q, got %q", v, got)
		}
	}
}

func TestReadFontNameDefaultsToSystem(t *testing.T) {
	got := ReadFontName("unknown")
	if got != DefaultFontName {
		t.Fatalf("expected %q, got %q", DefaultFontName, got)
	}
}

func TestReadFontNameAcceptsMonospace(t *testing.T) {
	got := ReadFontName("Monospace")
	if got != "Monospace" {
		t.Fatalf("expected Monospace, got %q", got)
	}
}

func TestDensityOptionsCount(t *testing.T) {
	if len(DensityOptions()) != 3 {
		t.Fatalf("expected 3 density options, got %d", len(DensityOptions()))
	}
}
