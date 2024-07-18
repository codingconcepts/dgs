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
| ${street_suffix} | berg |
| ${time_zone_region} | Europe/Kirov |
| ${verb_action} | play |
| ${connective_complaint} | e.g. |
| ${noun_collective_thing} | archipelago |
| ${pronoun_interrogative} | whom |
| ${quote} | "Messenger bag pork belly wayfarers." - Hellen Botsford |
| ${lorem_paragraph} | Dolorem ut placeat impedit nam reprehenderit sed nam natus tempora ab consequatur provident ducimus sapiente... |
| ${preposition_double} | up to |
| ${price} | 17.59 |
| ${celebrity_actor} | Mel Gibson |
| ${credit_card_exp} | 02/25 |
| ${file_mime_type} | text/plain |
| ${image_jpg} | [255 216 255 219 0 132 0 8 6 6 7 ...] |
| ${connective_time} | on another occasion |
| ${int32} | 1164719129 |
| ${pronoun} | I |
| ${month_string} | July |
| ${name_prefix} | Ms. |
| ${noun_collective_animal} | pod |
| ${noun} | cast |
| ${buzz_word} | process improvement |
| ${hour} | 3 |
| ${http_version} | HTTP/2.0 |
| ${job_title} | Agent |
| ${past_date} | 2024-07-17 23:54:23.96433 +0100 BST m=-39599.988705207 |
| ${uint64} | 3060866867892574108 |
| ${street_prefix} | North |
| ${uint16_hex} | 0x6241 |
| ${vegetable} | Parsnip |
| ${word} | explode |
| ${connective} | such as |
| ${farm_animal} | Cow |
| ${job_level} | Configuration |
| ${phone_formatted} | 436-128-1918 |
| ${dessert} | Grammie millers swedish apple pie |
| ${float64} | 0.828329343823015 |
| ${password} | f@Zfe@2FE fzsgC2-p809o TR |
| ${preposition_simple} | by |
| ${weekday} | Wednesday |
| ${connective_casual} | an upshot of |
| ${currency_long} | Falkland Islands (Malvinas) Pound |
| ${month} | 4 |
| ${noun_collective_people} | troop |
| ${error_database} | table migration failed |
| ${lorem_sentence} | Quidem velit distinctio expedita hic quibusdam repellat nesciunt quia eos quisquam qui qui fugit fugit ... |
| ${noun_countable} | bush |
| ${opera_user_agent} | Opera/9.29 (X11; Linux x86_64; en-US) Presto/2.10.200 Version/13.00 |
| ${adjective_possessive} | my |
| ${animal} | ant |
| ${app_version} | 2.5.7 |
| ${currency_short} | ILS |
| ${time_zone} | E. Africa Standard Time |
| ${url} | http://www.seniorcross-media.name/web-readiness/content |
| ${error} | database not initialized |
| ${nicecolors} | [#ffe181 #eee9e5 #fad3b2 #ffba7f #ff9c97] |
| ${pet_name} | Gary |
| ${pronoun_object} | me |
| ${product_category} | jewelry |
| ${ssn} | 651742334 |
| ${uuid} | 5ed012d2-ffb8-472f-9c29-dc02b82a9819 |
| ${error_grpc} | connection refused |
| ${future_date} | 2024-07-18 16:54:23.964482 +0100 BST m=+21600.011447126 |
| ${int8} | 43 |
| ${product_name} | Mixer Pure Water-Resistant |
| ${language} | Yiddish |
| ${name_suffix} | DDS |
| ${car_transmission_type} | Manual |
| ${emoji} | ✅ |
| ${http_status_code} | 304 |
| ${language_abbreviation} | kv |
| ${adverb_frequency_indefinite} | frequently |
| ${movie_genre} | Sci-Fi |
| ${verb} | was |
| ${minute} | 8 |
| ${phone} | 2007658555 |
| ${product_material} | paper |
| ${verb_helping} | was |
| ${adjective_demonstrative} | this |
| ${adjective_interrogative} | how |
| ${adverb_degree} | incredibly |
| ${adverb} | then |
| ${zip} | 89836 |
| ${car_model} | M5 |
| ${longitude} | -143.541965 |
| ${pronoun_demonstrative} | those |
| ${pronoun_relative} | whichever |
| ${adverb_manner} | rightfully |
| ${bitcoin_private_key} | 5KYgHSdgw4RQ32DyhoyUnrn5VFngjDyoJRmaZURWPHL1qMW2o4o |
| ${bool} | false |
| ${breakfast} | Blueberry bakery muffins |
| ${school} | Hawthorn Private University |
| ${adverb_time_indefinite} | yet |
| ${car_maker} | Citroen |
| ${company_suffix} | and Sons |
| ${file_extension} | html |
| ${safari_user_agent} | Mozilla/5.0 (Macintosh; PPC Mac OS X 10_7_9 rv:4.0; en-US) AppleWebKit/531.30.8 (KHTML, like Gecko) Version/5.1 Safari/531.30.8 |
| ${car_type} | Van |
| ${connective_examplify} | for instance |
| ${noun_uncountable} | equipment |
| ${pronoun_personal} | he |
| ${uint32_hex} | 0x2d9fd8c0 |
| ${uint8_hex} | 0x78 |
| ${book_title} | Sons and Lovers |
| ${hipster_word} | schlitz |
| ${latitude} | -83.517001 |
| ${state} | Oklahoma |
| ${lorem_word} | nobis |
| ${noun_concrete} | host |
| ${rgbcolor} | [31 189 107] |
| ${bs} | e-tailers |
| ${chrome_user_agent} | Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8) AppleWebKit/5321 (KHTML, like Gecko) Chrome/36.0.808.0 Mobile Safari/5321 |
| ${credit_card_number} | 6062827995210786 |
| ${error_runtime} | undefined has no such property 'length' |
| ${movie_name} | Finding Nemo |
| ${product_feature} | biometric |
| ${pronoun_reflective} | ourselves |
| ${safecolor} | white |
| ${dinner} | Jiffy punch |
| ${ipv6_address} | aad9:49b6:37bb:55e9:68a0:c33a:5284:20e4 |
| ${job_descriptor} | Principal |
| ${lunch} | Chocolate almond roca bar |
| ${time_zone_offset} | 3 |
| ${uint32} | 1840782908 |
| ${year} | 1932 |
| ${domain_name} | centralengage.info |
| ${http_status_code_simple} | 404 |
| ${int64} | 8829257814977890184 |
| ${programming_language} | Oberon |
| ${adjective_descriptive} | red |
| ${adverb_frequency_definite} | daily |
| ${connective_listing} | to conclude |
| ${credit_card_type} | Maestro |
| ${snack} | Delicious and simple fruit dip |
| ${verb_linking} | does |
| ${username} | Murphy5057 |
| ${float32} | 0.89757884 |
| ${http_method} | PUT |
| ${last_name} | Reynolds |
| ${time_zone_abv} | HST |
| ${pronoun_possessive} | mine |
| ${uint16} | 39488 |
| ${int16} | -17752 |
| ${nanosecond} | 426055826 |
| ${preposition} | along with |
| ${product_description} | Inspect it me instead while neck today Kyrgyz yet wad. |
| ${country} | Saint Barthélemy |
| ${day} | 26 |
| ${hobby} | Ghost hunting |
| ${phrase} | or something |
| ${mac_address} | 08:a4:47:44:a9:12 |
| ${street_number} | 67745 |
| ${uint128_hex} | 0x7a915faaccf77ad43f3e9f202a51032b |
| ${app_name} | Regimenthad |
| ${company} | xDayta |
| ${error_http_client} | im a teapot |
| ${hipster_sentence} | Loko loko jean shorts fashion axe wayfarers intelligentsia irony freegan waistcoat vinegar PBR&B .... |
| ${noun_abstract} | anger |
| ${uint8} | 103 |
| ${user_agent} | Mozilla/5.0 (Macintosh; PPC Mac OS X 10_8_9) AppleWebKit/5350 (KHTML, like Gecko) Chrome/40.0.861.0 Mobile Safari/5350 |
| ${adjective_proper} | Sri-Lankan |
| ${city} | Chula Vista |
| ${email} | mackenziecole@glover.com |
| ${name} | Enos Graham |
| ${color} | Cornsilk |
| ${firefox_user_agent} | Mozilla/5.0 (X11; Linux i686; rv:7.0) Gecko/1986-09-13 Firefox/37.0 |
| ${noun_common} | case |
| ${second} | 56 |
| ${company_slogan} | algorithm action-items Family, knowledge user. |
| ${hexcolor} | #bcbe72 |
| ${image_png} | [137 80 78 71 13 10 26 10 0 ...] |
| ${street_name} | Parkways |
| ${adjective_quantitative} | sufficient |
| ${adverb_time_definite} | now |
| ${app_author} | Zechariah Tillman |
| ${book_author} | Astrid Lindgren |
| ${error_http_server} | not implemented |
| ${gender} | male |
| ${ipv4_address} | 219.125.34.114 |
| ${isin} | GGHB1IW25R50 |
| ${bitcoin_address} | 3p9EOd05o4lJf9gu2VD99h7sy97 |
| ${book_genre} | Speculative |
| ${country_abr} | CI |
| ${credit_card_cvv} | 400 |
| ${time_zone_full} | (UTC+12:00) Magadan |
| ${uint64_hex} | 0xea07822251f8e6f1 |
| ${cusip} | SCLZ40FK0 |
| ${error_http} | invalid method |
| ${hipster_paragraph} | Quinoa ugh artisan organic wayfarers you probably haven't heard of them five dollar toast slow-carb ... |
| ${car_sport} | Jackie Joyner-Kersee |
| ${flipacoin} | Heads |
| ${state_abr} | ND |
| ${ach_routing} | 754697009 |
| ${adjective_indefinite} | some |
| ${adverb_place} | over |
| ${car_business} | Phil Knight |
| ${car_fuel_type} | Ethanol |
| ${first_name} | Keanu |
| ${question} | Gluten-free asymmetrical tacos PBR&B street 8-bit literally? |
| ${street} | 59013 Crescentfort |
| ${domain_suffix} | info |
| ${fruit} | Blackberry |
| ${preposition_compound} | in favor of |
| ${uint256_hex} | 0x9e473d233a81574126092e4eee878c616e795e4f62176353e088211d7f81bc83 |
| ${ach_account} | 309982025208 |
| ${adjective} | that |
| ${animal_type} | fish |
| ${date} | 1992-12-06 20:25:28.400810941 +0000 UTC |

### Todo

Performance

* Consider sorting data by primary key column(s) before inserting

Parity with [dg](https://github.com/codingconcepts/dg)

* each
* range
* match
* CSV generation