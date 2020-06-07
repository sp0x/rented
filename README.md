# rented
This project tracks apartment renting sites and waits for new properties to show up.  
Whenever a new property shows up you will be notified by Telegram.

### Configuration

Environment variables:
- `TELEGRAM_TOKEN` To use the telegram functionality you need to provide a token
- `INDEXER` The names of the indexers for property sites.

### Supported sites  
- Cityapartment.dk - `cityapartment`
  
You can add new sites in a few ways:
- Add a new YML file in the `./sites/` directory
- Add a new YML file in the `~./.rented/sites/` directory
- Add a new YML file in the sites directory of the project and rebuild it.
