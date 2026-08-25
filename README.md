# golol

Selector de objetos de League of Legends: una sola ruta, `/items`, al estilo de la pestaña All Items de la tienda. Go + HTMX, datos de [Data Dragon](https://developer.riotgames.com/docs/lol#data-dragon).

## Cómo arrancarlo

```bash
make run
```

Abre [http://localhost:8080/items](http://localhost:8080/items). El primer arranque descarga `item.json` del CDN (sin API key) y lo deja en `.cache/ddragon/`.

| Variable | Default | Qué hace |
|---|---|---|
| `ADDR` | `:8080` | Dirección de escucha |
| `DDRAGON_LOCALE` | `es_ES` | Locale de Data Dragon |
| `CACHE_DIR` | `.cache/ddragon` | Cache de `item.json` |

```bash
make test
```

## Filtros

- **Rol** (uno): Todos, Luchador, Tirador, Asesino, Mago, Tanque, Soporte. Data Dragon no trae clase; el rol se deriva de los `tags`.
- **Stats** (varios, AND): el objeto tiene que cumplir **todos** los checks marcados. Ejemplo: `/items?role=tank&stat=Armor&stat=SpellBlock`.

Solo se listan objetos comprables en la Grieta del Invocador.

golol isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties. Riot Games and all associated properties are trademarks or registered trademarks of Riot Games, Inc.
