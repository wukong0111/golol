# golol

Selector de objetos y campeones de League of Legends. Go + HTMX, datos de [Data Dragon](https://developer.riotgames.com/docs/lol#data-dragon).

- [`/items`](http://localhost:8080/items) — tienda de objetos, al estilo de la pestaña All Items.
- [`/champions`](http://localhost:8080/champions) — roster filtrable por rol, con splash y kit.

## Cómo arrancarlo

```bash
make run
```

Abre [http://localhost:8080/items](http://localhost:8080/items) o [http://localhost:8080/champions](http://localhost:8080/champions). El primer arranque descarga `item.json` y `championFull.json` del CDN (sin API key) y los deja en `.cache/ddragon/`.

| Variable | Default | Qué hace |
|---|---|---|
| `ADDR` | `:8080` | Dirección de escucha (gana a `PORT`) |
| `PORT` | — | Puerto que inyecta Railway |
| `DDRAGON_LOCALE` | `es_ES` | Locale de Data Dragon |
| `CACHE_DIR` | `.cache/ddragon` | Cache de `item.json` y `championFull.json` |
| `SHUTDOWN_TIMEOUT` | `15s` (o draining de Railway − 1s) | Tiempo máximo para terminar requests al apagar |

En SIGTERM/SIGINT el proceso deja `/health` en 503, deja de aceptar conexiones nuevas y espera a que acaben las que ya estaban en vuelo (`http.Server.Shutdown`). `railway.toml` configura el healthcheck y 15s de draining (Railway por defecto manda SIGKILL al instante).

```bash
make test
```

## Filtros

### Objetos (`/items`)

- **Rol** (uno): Todos, Luchador, Tirador, Asesino, Mago, Tanque, Soporte. Data Dragon no trae clase; el rol se deriva de los `tags`.
- **Stats** (varios, AND): el objeto tiene que cumplir **todos** los checks marcados. Ejemplo: `/items?role=tank&stat=Armor&stat=SpellBlock`.

Solo se listan objetos comprables en la Grieta del Invocador.

### Campeones (`/champions`)

- **Rol** (varios, OR): Luchador, Tirador, Asesino, Mago, Tanque, Soporte. Sin checks = todos. Un campeón entra si tiene **alguno** de los roles marcados. Ejemplo: `/champions?role=fighter&role=tank`.
- Al seleccionar un campeón se muestra el splash de la skin por defecto y las habilidades P/Q/W/E/R.

golol isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties. Riot Games and all associated properties are trademarks or registered trademarks of Riot Games, Inc.
