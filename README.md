# git-backed-cqrs

## An example of an event-sourced CQRS solution in GoLang

There's an old saw about how the only way to really learn something is to actually do it.

I've been following along with the various __CQRS__ projects from the Vancouver DDD/CQRS/ES Meetup, but this is the first one that really grabbed me.

The premise is relatively simple; perhaps I can contribute. If not I'll certainly learn some Git while tracking progress. [UPDATE: I thought 'git' felt really familiar but put this down to my heavy use of CVS back in the day. However, after getting pinged on a patch that I posted years ago and installing 'mercurial' to verify, it became obvious to me that 'git' and 'mercurial' seem very much the same sort of thing.]

#### Dependencies

As might seem obvious from the "git-backed" part, there is a run-time requirement that __git__ be installed.

#### To run

	go run main.go

#### Web interfaces

To create a Todo list:

	localhost:8080/createTodoList

To query the Todo lists:

	localhost:8080/Todolists
