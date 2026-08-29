# golol

Selector de objetos y campeones de League of Legends. Go + HTMX, datos de [Data Dragon](https://developer.riotgames.com/docs/lol#data-dragon).

- [`/items`](http://localhost:8080/items) — tienda de objetos, al estilo de la pestaña All Items.
- [`/champions`](http://localhost:8080/champions) — roster filtrable por rol, con splash y kit.
- [`/builds`](http://localhost:8080/builds) — creador de builds: campeón + 7 objetos, guardado en el navegador.

## Cómo arrancarlo

```bash
make run
```

Abre [http://localhost:8080/items](http://localhost:8080/items), [http://localhost:8080/champions](http://localhost:8080/champions) o [http://localhost:8080/builds](http://localhost:8080/builds). El primer arranque descarga `item.json` y `championFull.json` de Data Dragon (sin API key), y los dumps de Meraki para kits y clases de tienda, y los deja en `.cache/ddragon/`.

| Variable | Default | Qué hace |
|---|---|---|
| `ADDR` | `:8080` | Dirección de escucha (gana a `PORT`) |
| `PORT` | — | Puerto que inyecta Railway |
| `DDRAGON_LOCALE` | `es_ES` | Locale de Data Dragon |
| `CACHE_DIR` | `.cache/ddragon` | Cache de Data Dragon y dumps de Meraki |
| `SHUTDOWN_TIMEOUT` | `15s` (o draining de Railway − 1s) | Tiempo máximo para terminar requests al apagar |

En SIGTERM/SIGINT el proceso deja `/health` en 503, deja de aceptar conexiones nuevas y espera a que acaben las que ya estaban en vuelo (`http.Server.Shutdown`). `railway.toml` configura el healthcheck y 15s de draining (Railway por defecto manda SIGKILL al instante).

```bash
make test
```

## Filtros

### Objetos (`/items`)

- **Rol** (uno): Todos, Luchador, Tirador, Asesino, Mago, Tanque, Soporte. Las clases salen de `shop.tags` de Meraki (el menú del cliente). Si Meraki no trae el objeto, se infieren de los `tags` de Data Dragon.
- **Stats** (varios, AND): el objeto tiene que cumplir **todos** los checks marcados. Ejemplo: `/items?role=tank&stat=Armor&stat=SpellBlock`.

Solo se listan objetos comprables en la Grieta del Invocador. Data Dragon marca clones de otros modos (ARAM `32xxxx`, prismáticos de Arena `66xxxx`, starters del Abismo) como mapa 11; esos IDs se descartan.

### Campeones (`/champions`)

- **Rol** (varios, OR): Luchador, Tirador, Asesino, Mago, Tanque, Soporte. Sin checks = todos. Un campeón entra si tiene **alguno** de los roles marcados. Ejemplo: `/champions?role=fighter&role=tank`.
- Al seleccionar un campeón se muestra el splash de la skin por defecto y las habilidades P/Q/W/E/R.

### Builds (`/builds`)

- Dos selectores con buscador (campeones y objetos). Pulsa **Añadir build** para crear una colección: un hueco de campeón y siete de objetos (el extra de esta temporada para ADC).
- Al seleccionar una colección entras en modo edición: el campeón u objeto que elijas se añade a esa build. Pulsar un objeto de la colección lo quita.
- Las colecciones se guardan en `localStorage` como JSON (`golol.builds`).

golol isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties. Riot Games and all associated properties are trademarks or registered trademarks of Riot Games, Inc.
