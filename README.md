# rented
This project tracks apartment renting sites and waits for new properties to show up.  
Whenever a new property shows up you will be notified by Telegram.

### Configuration

Environment variables:
- `TELEGRAM_TOKEN` To use the telegram functionality you need to provide a token
- `INDEXER` The names of the indexers for property sites.

### Supported sites  
- Cityapartment.dk - `cityapartment`
In order to add a new site, add a new Yaml file inside the sites directory and rebuild.  
For now this project uses embedded site definitions only.