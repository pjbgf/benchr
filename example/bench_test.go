package bench

import (
	"log"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
)

var (
	localRepoDir string
	repoURL      = "https://github.com/go-git/go-git.git"
)

func TestMain(m *testing.M) {
	var err error
	localRepoDir, err = os.MkdirTemp("", "bench-test")
	if err != nil {
		log.Fatal(err)
	}

	_, err = git.PlainClone(localRepoDir, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	os.Exit(code)
}

func BenchmarkPlainOpen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := git.PlainOpen(localRepoDir)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStatus(b *testing.B) {
	r, err := git.PlainOpen(localRepoDir)
	if err != nil {
		b.Fatal(err)
	}

	wt, err := r.Worktree()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		wt.Status()
	}
}
