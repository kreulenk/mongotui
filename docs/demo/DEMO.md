# Demo

The files contained in this directory all have to do with the generation of the demo
gif that exists within the readme of this repository.

The demo gif is generated using the VHS gif recording tool which will need to be installed on your
system to generate the gif yourself. More information on VHS can be found here -- https://github.com/charmbracelet/vhs

## Setting Up DB for Demo gif generation

In order to generate a demo gif you will first need to initialize your database with the proper data that is
expected by the `demo.tape` script. The `onlineShop.products.json` and `onlineShop.users.json` files
contain the collections found in the gif that exist under the onlineShop database. Ensure that you have
an empty db besides these collections to record the gif. You can use the `mongoimport` tool to load
the data onto your mongo server. E.g.
```
mongoimport -d onlineShop -c products --jsonArray --maintainInsertionOrder mongodb://localhost docs/demo/onlineShop.products.json
mongoimport -d onlineShop -c users --jsonArray --maintainInsertionOrder mongodb://localhost docs/demo/onlineShop.users.json
```

Additionally, you will want to run the 'data generator' tool in this repository to generate the data necessary for
the pagination section of the demo. E.g. `go run tools/data-generator/main.go localhost`

# Creating Demo Gif

Once you have the proper data in place, run `make demo-gif` from the root of this repository.

You should now have a newly generated demo.gif!