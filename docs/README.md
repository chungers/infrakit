
# README (docs authoring and publishing)

### Publishing the Beta docs from private-invites

See the README at https://github.com/docker/private-invites.

### Generating the docs locally

To generate documentation:
```
cd docs
make
```
Then browse to `http://localhost:8000/` to look at the doc.

In development mode, if you run Docker for Mac, you can leverage the dev target, that will auto refresh the site when you modify any file on your Mac (Thank you Cambridge team!).
```
make dev
```

### Hugo, themes, and build commands

Write Documentation in docs/content/, in markdown format.

We use [Hugo](https://gohugo.io/) and the regular [Docker documentation tooling and theme](https://hub.docker.com/r/docs/base/). This will allow us to be included in regular docs.docker.com when we are ready to ship, and inherit the theme changes the docs team is planning in the next few weeks, while having control on our own theme. The theme in editions is customizable: add Hugo theme material in docs/layouts or other Hugo theme standard directories, and it will be picked up by the Dockerfile for site generation.

The menubar is dynamically generated from the [front matter](https://gohugo.io/content/front-matter/) section at the top of each .md file. ```[menu.editions]``` annotation will assign that page in the menu. ```weight = 1``` allows you to affect menu items sort order. More details on how to specify hierarchical menubars at [Tips on Hugo metadata and menu positioning](https://github.com/docker/machine/tree/master/docs#tips-on-hugo-metadata-and-menu-positioning).

### Legacy publishing workflow on Azure (no longer used)

```
make DOCS_EXPORT=public docs-export
```
will generate static html docs in the public directory in current directory.
```
make clean DOCS_EXPORT=public
```
to clean it up.
