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

func TestSortGithubReposByLastCommit(t *testing.T) {
	oldest := time.Date(2026, time.November, 10, 12, 0, 0, 0, time.UTC)
	newest := time.Date(2026, time.November, 27, 12, 0, 0, 0, time.UTC)
	middle := time.Date(2026, time.November, 20, 12, 0, 0, 0, time.UTC)

	repos := []githubRepo{
		{FullName: "owner/oldest", LastCommit: oldest},
		{FullName: "owner/newest", LastCommit: newest},
		{FullName: "owner/middle", LastCommit: middle},
	}

	sortGithubReposByLastCommit(repos)

	want := []string{"owner/newest", "owner/middle", "owner/oldest"}
	for i := range want {
		if repos[i].FullName != want[i] {
			t.Fatalf("repos[%d] = %q, want %q", i, repos[i].FullName, want[i])
		}
	}
}
