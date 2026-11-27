package glance

import (
	"testing"
	"time"
)

func TestGithubRepoLastActivityPrefersPushedAt(t *testing.T) {
	repo := githubUserRepoResponseJson{
		PushedAt:  "2026-11-25T11:49:54Z",
		UpdatedAt: "2026-11-12T12:46:27Z",
	}

	want := time.Date(2026, time.November, 25, 11, 49, 54, 0, time.UTC)
	if got := githubRepoLastActivity(repo); !got.Equal(want) {
		t.Fatalf("githubRepoLastActivity() = %s, want %s", got, want)
	}
}

func TestGithubRepoLastActivityFallsBackToUpdatedAt(t *testing.T) {
	repo := githubUserRepoResponseJson{
		UpdatedAt: "2026-11-12T12:46:27Z",
	}

	want := time.Date(2026, time.November, 12, 12, 46, 27, 0, time.UTC)
	if got := githubRepoLastActivity(repo); !got.Equal(want) {
		t.Fatalf("githubRepoLastActivity() = %s, want %s", got, want)
	}
}

func TestGithubRepoLastActivityFallsBackWhenPushedAtIsInvalid(t *testing.T) {
	repo := githubUserRepoResponseJson{
		PushedAt:  "invalid",
		UpdatedAt: "2026-11-12T12:46:27Z",
	}

	want := time.Date(2026, time.November, 12, 12, 46, 27, 0, time.UTC)
	if got := githubRepoLastActivity(repo); !got.Equal(want) {
		t.Fatalf("githubRepoLastActivity() = %s, want %s", got, want)
	}
}

func TestGithubRepoLastActivityPreservesInstantWithOffset(t *testing.T) {
	repo := githubUserRepoResponseJson{
		PushedAt: "2026-11-25T12:49:54+01:00",
	}

	want := time.Date(2026, time.November, 25, 11, 49, 54, 0, time.UTC)
	if got := githubRepoLastActivity(repo); !got.Equal(want) {
		t.Fatalf("githubRepoLastActivity() = %s, want %s", got, want)
	}
}
