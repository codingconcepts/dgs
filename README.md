# dgs
A streaming version of dg, which writes data directly to a database without any kind of buffering.

### Installation

Find the release that matches your architecture on the [releases](https://github.com/codingconcepts/dgs/releases) page.

Download the tar, extract the executable, and move it into your PATH. For example:

```sh
tar -xvf dgs_0.0.1_macos_amd64.tar.gz
```

### Usage

dgs uses cobra for managing commands, of which there are currently 2:

```
Usage:
  dgs gen [command]

Available Commands:
  config      Generate the config file for a given database schema
  data        Generate relational data
```

### Generate config

If familiar with dgs configuration, you may prefer to hand-roll your dgs configs. However, if you'd prefer to use dgs itself to generate the configuration for you, you can use `dgs gen config` to generate a configuration file.

Note that this tool will sort the tables in the config file in topological order (guaranteeing that tables with a reference to another table will be generated after the table they depend reference).

Generate config file with default row counts

```sh
dgs gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public > examples/e-commerce/config.yaml
```

Generate config file with custom row counts (tables without a row-count will receive the default row count)

```sh
dgs gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public \
--row-count member:100000 \
--row-count product:10000 \
--row-count purchase:200000 \
--row-count purchase_line:400000 > examples/e-commerce/config.yaml
```

### Generate data

Once you have a dgs config file, you can generate data.

```sh
dgs gen data \
--config examples/e-commerce/config.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 4 \
--batch 10000
```

### Data types

##### Value

Generate a random value for a column, using any of the [Random generator functions](#random-generator-functions).

```yaml
- name: id
  value: ${uuid}
```

##### Range

Generate a random value between a minimum and maximum value.

```yaml
- name: date_of_birth
  range: timestamp
  props:
    min: 2014-07-20T01:00:00+01:00
    max: 2024-07-17T01:00:00+01:00
    format: "2006-01-02T15:04:05Z"

- name: average_session_duration
  range: interval
  props:
    min: 10m
    max: 3h

- name: percentage_complete
  range: int
  props:
    min: 1
    max: 100

- name: price
  range: float
  props:
    min: 1.99
    max: 99.99

- name: password
  range: bytes
  props:
    min: 1
    max: 1000

- name: location
  range: point
  props:
    lat: 51.04284752235447
    lon: -0.8911379425829405
    distance_km: 100
```

##### Inc

Generate a monotonically incrementing value for a column, starting from a given number.

```yaml
- name: id
  inc: 1
```

##### Set

Generate a random value from a set of available values.

```yaml
- name: user_type
  set: [regular, read_only, admin]
```

##### Ref

Reference a column value generated for a previous table by referencing it by `table_name.column_name`.

```yaml
- name: order_id
  ref: order.id
```

##### Array

Generate an array of values using a given a [Random generator function](#random-generator-functions).

```yaml
- name: favourite_fruits
  array: ${fruit}
  props:
    min: 1
    max: 10
```

### Random generator functions

| Fake function | Example |
| ------------- | ------- |
| ${ach_account} | 913329610684 |
| ${ach_routing} | 954052301 |
| ${adjective_demonstrative} | here |
| ${adjective_descriptive} | expensive |
| ${adjective_indefinite} | anyone |
| ${adjective_interrogative} | what |
| ${adjective_possessive} | its |
| ${adjective_proper} | Polynesian |
| ${adjective_quantitative} | substantial |
| ${adjective} | Senegalese |
| ${adverb_degree} | fully |
| ${adverb_frequency_definite} | quarterly |
| ${adverb_frequency_indefinite} | rarely |
| ${adverb_manner} | unexpectedly |
| ${adverb_place} | out |
| ${adverb_time_definite} | tomorrow |
| ${adverb_time_indefinite} | before |
| ${adverb} | often |
| ${animal_type} | invertebrates |
| ${animal} | minnow |
| ${app_author} | ASC Partners |
| ${app_name} | Yellowjacketbathe |
| ${app_version} | 2.7.19 |
| ${bitcoin_address} | 344vpjQxLmbV414cL89qypgJ129I2rphSi |
| ${bitcoin_private_key} | 5KpHhVaqcShwHYwYkvCmeFXM5H6B2JRdkqjHBaL7DfTEqE6w212 |
| ${book_author} | William Shakespeare |
| ${book_genre} | Mystery |
| ${book_title} | Zorba the Greek |
| ${bool} | false |
| ${breakfast} | Egg flowers |
| ${bs} | extensible |
| ${buzz_word} | analyzer |
| ${car_business} | Adam Weitsman |
| ${car_fuel_type} | Electric |
| ${car_maker} | Cadillac |
| ${car_model} | Highlander 2wd |
| ${car_sport} | Andy Murray |
| ${car_transmission_type} | Manual |
| ${car_type} | Pickup truck |
| ${celebrity_actor} | Peter Jackson |
| ${chrome_user_agent} | Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_7_4) AppleWebKit/5320 (KHTML, like Gecko) Chrome/37.0.898.0 Mobile Safari/5320 |
| ${city} | Boise |
| ${color} | ForestGreen |
| ${company_slogan} | 24 hour Care, leverage Expression. |
| ${company_suffix} | Inc |
| ${company} | Informatica |
| ${connective_casual} | nevertheless |
| ${connective_complaint} | in other words |
| ${connective_examplify} | provided that |
| ${connective_listing} | for one thing |
| ${connective_time} | at this point |
| ${connective} | for instance |
| ${country_abr} | HR |
| ${country} | Turks and Caicos Islands |
| ${credit_card_cvv} | 759 |
| ${credit_card_exp} | 01/27 |
| ${credit_card_number} | 598987466642233434 |
| ${credit_card_type} | Hipercard |
| ${currency_long} | Sri Lanka Rupee |
| ${currency_short} | SDG |
| ${cusip} | 0RLFMNX21 |
| ${date} | 1924-06-04 14:29:12.804889916 +0000 UTC |
| ${day} | 3 |
| ${dessert} | Old school deja vu chocolate peanut butter squares |
| ${dinner} | Savory pita chips |
| ${domain_name} | internationaltarget.org |
| ${domain_suffix} | net |
| ${email} | lawrencelabadie@huel.com |
| ${emoji} | 🧯 |
| ${error_database} | database connection error |
| ${error_grpc} | connection is shut down |
| ${error_http_client} | payload too large |
| ${error_http_server} | bad gateway |
| ${error_http} | invalid method |
| ${error_runtime} | not enough arguments |
| ${error} | [object Object] |
| ${farm_animal} | Sheep |
| ${file_extension} | avi |
| ${file_mime_type} | image/bmp |
| ${firefox_user_agent} | Mozilla/5.0 (Windows NT 5.1; en-US; rv:1.9.1.20) Gecko/1985-10-11 Firefox/36.0 |
| ${first_name} | Pansy |
| ${flipacoin} | Tails |
| ${float32} | 0.5662413 |
| ${float64} | 0.33995204123600065 |
| ${fruit} | Rambutan |
| ${future_date} | 2024-07-18 16:10:46.558835 +0100 BST m=+18000.000711668 |
| ${gender} | female |
| ${hexcolor} | #3edf1e |
| ${hipster_paragraph} | Fixie goth pitchfork master messenger bag scenester post-ironic ... |
| ${hipster_sentence} | Lomo authentic humblebrag green juice butcher marfa brooklyn waistcoat ... |
| ${hipster_word} | selvage |
| ${hobby} | Homebrewing |
| ${hour} | 9 |
| ${http_method} | HEAD |
| ${http_status_code_simple} | 200 |
| ${http_status_code} | 502 |
| ${http_version} | HTTP/2.0 |
| ${image_jpg} | [255 216 255 219 0 132 0 8 6 6 7 6 5 8 7 ...] |
| ${image_png} | [137 80 78 71 13 10 26 10 0 0 0 13 73 72 ...] |
| ${int16} | -14249 |
| ${int32} | 1398589957 |
| ${int64} | 3609435043382311606 |
| ${int8} | -93 |
| ${ipv4_address} | 240.242.187.215 |
| ${ipv6_address} | d32c:ebb5:1e18:2620:40cc:f49b:3027:5459 |
| ${isin} | KWVTK4GLRO92 |
| ${job_descriptor} | Customer |
| ${job_level} | Applications |
| ${job_title} | Representative |
| ${language_abbreviation} | wa |
| ${language} | Macedonian |
| ${last_name} | Schumm |
| ${latitude} | 78.05631 |
| ${longitude} | -46.346742 |
| ${lorem_paragraph} | Perspiciatis veritatis commodi sed nam ex recusandae consequuntur ... |
| ${lorem_sentence} | Sequi eos quia molestias est doloribus at consequatur deleniti ... |
| ${lorem_word} | iure |
| ${lunch} | Salata marouli romaine lettuce salad |
| ${mac_address} | 24:1f:13:fd:6d:d4 |
| ${minute} | 34 |
| ${month_string} | July |
| ${month} | 10 |
| ${movie_genre} | Mystery |
| ${movie_name} | Intouchables |
| ${name_prefix} | Ms. |
| ${name_suffix} | V |
| ${name} | Stefanie Koelpin |
| ${nanosecond} | 285275999 |
| ${nicecolors} | [#a3a948 #edb92e #f85931 #ce1836 #009989] |
| ${noun_abstract} | care |
| ${noun_collective_animal} | exaltation |
| ${noun_collective_people} | posse |
| ${noun_collective_thing} | archipelago |
| ${noun_common} | hand |
| ${noun_concrete} | earrings |
| ${noun_countable} | holiday |
| ${noun_uncountable} | yoga |
| ${noun} | dynasty |
| ${opera_user_agent} | Opera/8.21 (Windows 98; en-US) Presto/2.12.257 Version/13.00 |
| ${password} | ZqM1A.53oQ NpK1H0F!3r95*! |
| ${past_date} | 2024-07-18 09:10:46.565036 +0100 BST m=-7199.993087291 |
| ${pet_name} | Bark Twain |
| ${phone_formatted} | 463.696.7163 |
| ${phone} | 3602152186 |
| ${phrase} | that's what I'm talking about |
| ${preposition_compound} | up to |
| ${preposition_double} | next to |
| ${preposition_simple} | till |
| ${preposition} | out of |
| ${price} | 40.65 |
| ${product_category} | tools and hardware |
| ${product_description} | Frequently place oops how way. |
| ${product_feature} | multi-functional |
| ${product_material} | silver |
| ${product_name} | Ultra-Lightweight Printer Fresh |
| ${programming_language} | BeanShell |
| ${pronoun_demonstrative} | that |
| ${pronoun_interrogative} | when |
| ${pronoun_object} | me |
| ${pronoun_personal} | she |
| ${pronoun_possessive} | ours |
| ${pronoun_reflective} | myself |
| ${pronoun_relative} | whose |
| ${pronoun} | as |
| ${question} | Microdosing tumblr organic? |
| ${quote} | "XOXO occupy tattooed meggings whatever drinking organic." - Madelynn Gutkowski |
| ${rgbcolor} | [174 224 122] |
| ${safari_user_agent} | Mozilla/5.0 (Windows; U; Windows NT 6.1) AppleWebKit/535.24.5 (KHTML, like Gecko) Version/4.1 Safari/535.24.5 |
| ${safecolor} | gray |
| ${school} | Harborview Private High School |
| ${second} | 25 |
| ${snack} | Spicy roasted butternut seeds pumpkin seeds |
| ${ssn} | 508233622 |
| ${state_abr} | FL |
| ${state} | North Dakota |
| ${street_name} | Lodge |
| ${street_number} | 6725 |
| ${street_prefix} | West |
| ${street_suffix} | mouth |
| ${street} | 615 New Courtfurt |
| ${time_zone_abv} | KST |
| ${time_zone_full} | (UTC-12:00) International Date Line West |
| ${time_zone_offset} | -3 |
| ${time_zone_region} | America/Mendoza |
| ${time_zone} | Paraguay Standard Time |
| ${uint128_hex} | 0xb04af94cb34cc417eba86daeee1e1326 |
| ${uint16_hex} | 0x64e7 |
| ${uint16} | 30988 |
| ${uint256_hex} | 0x8167bc9db8d22679668c7da369b2cad94a6ac979d6c013f2846296016e0cb92b |
| ${uint32_hex} | 0xe70e5724 |
| ${uint32} | 3437344718 |
| ${uint64_hex} | 0x19715e3d1d09a53e |
| ${uint64} | 16119066877620660254 |
| ${uint8_hex} | 0x81 |
| ${uint8} | 89 |
| ${url} | https://www.internationalsupply-chains.biz/unleash/cross-platform |
| ${user_agent} | Mozilla/5.0 (Macintosh; U; PPC Mac OS X 10_9_1 rv:7.0; en-US) AppleWebKit/535.30.2 (KHTML, like Gecko) Version/4.0 Safari/535.30.2 |
| ${username} | Williamson4993 |
| ${uuid} | 718396a8-145b-4522-bc2c-f1a572645cd9 |
| ${vegetable} | Snow Peas |
| ${verb_action} | eat |
| ${verb_helping} | should |
| ${verb_linking} | are |
| ${verb} | turn |
| ${weekday} | Wednesday |
| ${word} | infrequently |
| ${year} | 1914 |
| ${zip} | 80752 |

### Todo

Performance

* Consider sorting data by primary key column(s) before inserting

Parity with [dg](https://github.com/codingconcepts/dg)

* each
* range
* match
* CSV generation