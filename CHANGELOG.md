# Changelog

## feat/docker-container-column

Drilling into Docker processes now reveals which containers own each connection (#12).
A new "Container" column appears showing container name, image, and port mapping
(e.g., `nginx (nginx:latest) 8080→80`) — powered by the Docker Engine API with
async resolution and port-keyed caching. The column is fully sortable and degrades
gracefully when Docker is unavailable. New `internal/docker` package with `Resolver`
interface and comprehensive test coverage across model, resolver, UI, and integration layers.
