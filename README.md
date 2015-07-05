This repository contains small tools that I use.

* **bmk**: A tool to "bookmark" directories for fast access. Import bmk.bashrc in your .bashrc and use bcd to change directories.
* **findunversioned**: A tool to find directories that are not contained in any version controlled repository. This does **not** find subdirectories or files within version controlled repositories that are not being tracked. Good for running under a personal src/ dir for instance.
* **pipesplit**: A tool that will chunk data for you. This is a bit quirky as was designed to meet the needs of my personal backup system. It can optionally create SHA-256 hashes of files.
* **simpleserv**: Serves the directory under an http server.
* **srcsearch**: Similar to bcd, but rather than define bookmarks, set SRCSEARCHROOT in your environment and use srcsearch to find a directory with that name. Can also take partial paths (e.g. tools or uluyol/tools), but the name of the directory must be complete. srcsearch just returns a path, but a helper function scd defined in scd.bash will print the path and change directories for you.

Under bin/:

* **backme**: My personal incremental, encrypted backup script using rdup.
* **lsapps**: Lists applications in OSX.
* **pdfcat**: Concatenate PDF files.
* **startacme**: Starts acme while properly setting up the environment (allows for multiple windows).
