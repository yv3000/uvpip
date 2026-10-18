package main

import (
	"testing"
)

func BenchmarkTranslateToUV(b *testing.B) {
	args := []string{"install", "requests", "--upgrade", "-r", "requirements.txt"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := translateToUV(args)
		if len(res) == 0 {
			b.Fatal("unexpected empty translation")
		}
	}
}

func BenchmarkValidateArgs(b *testing.B) {
	args := []string{"install", "requests", "--system"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := ValidateArgs(args); err != nil {
			b.Fatalf("ValidateArgs failed: %v", err)
		}
	}
}

func BenchmarkCanonicalValidationSchema(b *testing.B) {
	args := []string{"install", "requests", "--upgrade"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := CanonicalValidationSchema.Validate(args); err != nil {
			b.Fatalf("CanonicalValidationSchema failed: %v", err)
		}
	}
}

func BenchmarkUVCandidatePaths(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		paths := uvCandidatePaths()
		if len(paths) == 0 {
			b.Fatal("expected candidate paths")
		}
	}
}
