# blockchain-data-aggregator

Take-Home Exercise: Blockchain Data Aggregator for Marketplace Analytics

## Structure

`rates` module contains daily batch job logic for extracting exchange rates data.

`bigquery` module contains logic for creating, populating and calculating result table in BigQuery.

`docs` directory contains diagrams.

`scripts` directory contains miscellaneous Bash scripts.

## Usage

To run the pipeline please follow the steps outlined below. 
Instructions assume that you have already created a Google Cloud project to use for Cloud Storage and BigQuery services.

## Cloud Storage Setup

1. Create Cloud Storage bucket.

2. Upload `sample_data.csv` to the bucket under `/raw/events` directory.

## Local Setup

1. [Install and initialize gcloud CLI](https://cloud.google.com/sdk/docs/install)

2. Provide user credentials to [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials#personal):
   ```bash
   gcloud auth application-default login
   ```

3. Obtain [CoinAPI](https://www.coinapi.io/get-free-api-key?email=) API key. After completing the form you should receive it in your email.
   _P.S. Make sure to select at least one checkbox in the form - turns out it is required even though there is no indication of that._

   I decided to go for CoinAPI instead of CoinGecko for several of reasons: 
   1. CoinGecko's exchange rates endpoint doesn't give you access to USD rates.
   2. It returns rates for a very limited number of coins.
   3. There is no access to historical data.

## Running the pipeline

There are 2 options for running the logic on your machine:
- Clone the repo and run the code using `go run`.
- Download and run pre-built binaries from the [latest release](https://github.com/mfridrikhson/blockchain-data-aggregator/releases).

### Running using `go run`

1. Install go1.23.0 and make sure your development environment is using that version.

2. `cd` into `/rates` module and run following command to populate raw exchange rates data for 2024-04-15. 
   ```bash
   go run . --bucketName your-bucket-name --apiKey coinapi-api-key --date 2024-04-15
   ```
   Repeat for other dates to collect more exchange rates data.

3. `cd` into `/bigquery` module and run following command to create and populate BigQuery tables:
   ```bash
   go run . \
      --projectId your-project-id\
      --sourceBucketId your-bucket-name\
      --datasetId data\
      --eventsTableId events\
      --ratesTableId rates\
      --resultTableId result
   ```

4. Pipeline execution complete. You can navigate to your BigQuery console and check the result table.

### Running using binaries

1. Download `rates` and `bigquery` binaries matching your architecture from the latest release's files.
2. Run `rates`:
   ```bash
   ./rates-os-arch --bucketName your-bucket-name --apiKey coinapi-api-key --date 2024-04-15
   ```
3. Run `bigquery`:
   ```bash
   ./bigquery-os-arch \
      --projectId your-project-id\
      --sourceBucketId your-bucket-name\
      --datasetId data\
      --eventsTableId events\
      --ratesTableId rates\
      --resultTableId result
   ```

## Data

Data Flow:

![Data Flow Diagram](/docs/data-flow.jpg)

### Table Schemas

Events table:

| Column | Type |
| --- | --- |
| app | STRING |
| ts | TIMESTAMP |
| event | STRING |
| project_id | INTEGER |
| source | STRING |
| ident | INTEGER |
| user_id | STRING |
| session_id | STRING |
| country | STRING |
| device_type | STRING |
| device_os | STRING |
| device_os_ver | STRING |
| device_browser | STRING |
| device_browser_ver | STRING |
| props | JSON |
| nums | JSON |

Rates table:

| Column | Type |
| --- | --- |
| date | DATE |
| rate | FLOAT |
| symbol | STRING |

Result table:

| Column | Type |
| --- | --- |
| date | DATE |
| project_id | INTEGER |
| num_transactions | INTEGER |
| total_volume_usd | FLOAT |

## Production Usage

All the logic presented in this repository is ready to be used in a production environment. In a production environment `rates` module could be used to run on a schedule (using Cloud Scheduler, Airflow or similar service) to collect daily exchange rates. We could also schedule the `bigquery` module in a similar fashion, however I don't see much value in doing so using Go code as BigQuery provides tools for running [scheduled data transfers](https://cloud.google.com/bigquery/docs/cloud-storage-transfer) using which we can append new data to the existing tables.

Possible architecture of the pipeline:

![Possible Architecture](/docs/possible-architecture.jpg)