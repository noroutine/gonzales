package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"strings"

	"github.com/noroutine/bonjour"
	"github.com/noroutine/gonzales/protocol"
	"github.com/noroutine/gonzales/cli"

	"github.com/reusee/mmh3"

	"golang.org/x/oauth2"
	"github.com/google/go-github/github"
)

const version = "0.0.1"
const description = "Gonzales " + version
const serviceType = "_gonzales._tcp"
const domain = "local."
const servicePort = 9999

type GitHubData struct {
	Token string
	Username string
}


func serviceList() {
    resolver, err := bonjour.NewResolver(nil)
    if err != nil {
        log.Println("Failed to initialize resolver:", err.Error())
   	    return
    }

    results := make(chan *bonjour.ServiceEntry)
    err = resolver.Browse(serviceType, domain, results)

    if err != nil {
        log.Println("Failed to browse:", err.Error())
        return
    }

L:
    for {
    	select {
    	case e := <- results:
    		fmt.Printf("%s @ %s (%v)\n", e.Instance, e.HostName, e.AddrIPv4)
    	case <- time.After(5 * time.Second):
    		break L
    	}
    }
}

func serviceRegister(name string) {
	// Run registration (blocking call)
    _, err := bonjour.Register(name, serviceType, "", servicePort, []string{"txtv=1", "app=test"}, nil)
    if err != nil {
        log.Fatalln(err.Error())
    }
    log.Printf("Registered")
}

func main() {

	port := os.Getenv("PORT")
	name := "gonzales Master"

	if port == "" {
		port = fmt.Sprintf("%d", servicePort)
	}

	client := &protocol.Client{ ":" + port, name }
	go client.Serve()

	repl := cli.New()
	repl.Description = description
	repl.Prompt = name + "> "

	go func() {
		for s := range repl.Signals {
			if s == os.Interrupt {
		    	log.Printf("Interrupted")
				os.Exit(0)
			}
		}
	}()

	ghData := &GitHubData{os.Getenv("GITHUB_ACCESS_TOKEN"), os.Getenv("GITHUB_USERNAME")}

	repl.EmptyHandler = func() {		
		fmt.Println("Feeling lost? Try 'help'")
		repl.EmptyHandler = nil
	}
	
	repl.Register("list", func(args []string) {
		serviceList()
	})

	repl.Register("register", func(args []string) {
		serviceRegister(name)
	})

	repl.Register("help", func(args []string) {
		fmt.Printf("Commands: %s\n", strings.Join(repl.GetKnownCommands(), ", "))
	})

	repl.Register("token", func(args []string) {
		token := ""
		if len(args) > 0 {
			token = args[0]
		}

		ghData.Token = token
	})

	repl.Register("username", func(args []string) {
		username := ""
		if len(args) > 0 {
			username = args[0]
		}
		ghData.Username = username
	})

	repl.Register("info", func(args []string) {
		fmt.Println("Token", ghData.Token)
		fmt.Println("User", ghData.Username)
	})

	repl.Register("organizations", func(args []string) {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ghData.Token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		gh := github.NewClient(tc)

		orgs, _, err := gh.Organizations.List(ghData.Username, nil)

		if err == nil {

			for _, org := range orgs {
				fmt.Println(*org.Login)
			}
		} else {
			fmt.Println(err)
		}
	})

	repl.Register("repositories", func(args []string) {		
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ghData.Token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		gh := github.NewClient(tc)

		namespace := ghData.Username
		var repos []github.Repository
		var err error

		if len(args) > 0 {
			repos, _, err = gh.Repositories.ListByOrg(args[0], nil)
			namespace = args[0]
		} else {
			repos, _, err = gh.Repositories.List("", nil)
		}

		if err == nil {
			for _, repo := range repos {
				fmt.Println(namespace + "/" + *repo.Name)
			}
		} else {
			fmt.Println(err)
		}
	})

	repl.Register("mmh3", func(args []string) {
		key := ""
		if len(args) > 0 {
			key = args[0]
		}

		fmt.Printf("murmur3(\"%s\") = %x\n", key, mmh3.Sum128([]byte(key)))
	})

	repl.Register("name", func(args []string) {
		if len(args) > 0 {
			name = args[0]
			fmt.Println("You are now", name)
			repl.Prompt = name + "> "
			client.PlayerID = name
		} else {
			fmt.Println(name)
		}
	})

	repl.Serve()
}
