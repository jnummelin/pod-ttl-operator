workflow "Basic test" {
  on = "push"
  resolves = ["Lint", "Test"]
}

action "Lint" {
  uses = "docker://docker.io/golang:1.11.4"
  runs = "go fmt"
}

action "Test" {
  needs = ["Lint"]
  uses = "docker://docker.io/golang:1.11.4"
  runs = "go test"
}
