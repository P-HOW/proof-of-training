# Proof-of-Training (POT)

![Go](https://img.shields.io/badge/Go-v1.20+-blue.svg)
[![Build Status](https://travis-ci.org/anfederico/clairvoyant.svg?branch=master)](https://travis-ci.org/anfederico/clairvoyant)
![Dependencies](https://img.shields.io/badge/dependencies-up%20to%20date-brightgreen.svg)
[![GitHub Issues](https://img.shields.io/github/issues/P-HOW/proof-of-training.svg)](https://github.com/P-HOW/proof-of-training/issues)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

:star: Star us on GitHub — it motivates us a lot!

Golang execution layer implementation of the decentralized training network using proof of training (POT).

![DTN](https://github.com/P-HOW/proof-of-training/blob/master/img/dtn.jpg?raw=true)

## Table Of Content

- [Layer-1 Implementation(L1)](#layer-1-implementation)
    - [Practical Byzantine Fault Tolerance (PBFT)](#pbft-package)
    - [(Recommended) Full Practical Byzantine Fault Tolerance (FPBFT)](#fpbft-package)
    - [TER](#ter-extension)
- [TYPO3 setup](#typo3-setup)
    - [Database setup](#database-setup)
    - [Security](#security)
- [Page setup](#page-setup)
    - [Download the Aimeos Page Tree t3d file](#download-the-aimeos-page-tree-t3d-file)
    - [Go to the Import View](#go-to-the-import-view)
    - [Upload the page tree file](#upload-the-page-tree-file)
    - [Go to the import view](#go-to-the-import-view)
    - [Import the page tree](#import-the-page-tree)
    - [SEO-friendly URLs](#seo-friendly-urls)
- [License](#license)
- [Links](#links)

## Layer-1 Implementation

This section focuses on two fundamental aspects of our Layer-1 (L1) implementation, namely:

- **Global Ledger Maintenance and Synchronization** 
- Multisignature Coordination

Our L1 implementation uses the [Practical Byzantine Fault Tolerance (PBFT)](https://pmg.csail.mit.edu/papers/osdi99.pdf) protocol, which ensures consensus among nodes in a distributed network, even in the presence of malicious nodes or if certain nodes fail.


### PBFT Package

The PBFT package ensures the uniformity of our global ledger by facilitating consensus among nodes on the set of transactions to be added. These approved transactions are synced to the global transaction pool, from which they are used to update the global ledger, ensuring data consistency across the network.

#### Functionality Implemented:
> PBFT formula: n >= 3f + 1 where n is the total number of nodes in the entire network, and f is the maximum number of malicious or faulty nodes allowed.

The data from client input to receiving replies from the nodes is divided into 5 steps:

1. The client sends request information to the primary node.
2. After the primary node N0 receives the request from the client, it extracts the main information from the request data and sends a preprepare to the other nodes.
3. The secondary nodes receive the preprepare from the primary node, firstly using the primary node's public key for signature authentication, then hash the message (message digest, to reduce the size of the information transmitted in the network), and broadcast a prepare to other nodes.
4. When a node receives 2f prepare information (including itself), and all signature verifications pass, it can proceed to the commit step, broadcasting a commit to all other nodes in the network.
5. When a node receives 2f+1 commit information (including itself), and all signature verifications pass, it can store the message locally, and return a reply message to the client.
   
![DTN](https://github.com/P-HOW/proof-of-training/blob/master/img/PBFTflow.png?raw=true)

To spawn a network applying PBFT with `numNodes` nodes, and to synchronize 
the transaction pool with transactions encoded in `data`, 
it is recommended to use the following function. This will help 
analyze the time needed for `data` synchronization.

Field | Data Types | Recommended Value
----  |------------| ----------
numNodes  | int        | 5-100
data  | string     | -
clientAddr  | string     | "127.0.0.1:8888"

```go
synctime := genPBFTSynchronize(numNodes, data, clientAddr)
```
> Note: The current implementations in the pbft package may contain 
> potential race conditions, potentially leading to non-terminating 
> execution. It is essential to implement a timing mechanism when using 
> this function for efficient operation.

#### pbft_test.go
```go
package pbft

import (
  "strconv"
  "testing"
)

func TestAddAndGetMessage(t *testing.T) {
  var clientAddr = "127.0.0.1:8888"
  var data = "transactions to be synchronized"
  var numNodes = 8
  sync_time := genPBFTSynchronize(numNodes, data, clientAddr)
  s := strconv.FormatFloat(sync_time, 'f', -1, 64)
  println("It takes " + s + " seconds to synchronize the transactions to the global ledger")
}
```
#### output
```text
=== RUN   TestAddAndGetMessage
the public and private key directory has not been generated yet, generating public and private keys...
RSA public and private keys have been generated for the nodes.
initiating client...
{"Content":"transactions to be synchronized","ID":2818938398,"Timestamp":1686148966474242384,"ClientAddr":"127.0.0.1:8888"}
The primary node has received a request from the client...
The request has been stored in the temporary message pool.
Broadcasting PrePrepare to other nodes...
PrePrepare broadcast completed.
It takes 0.032963246 seconds to synchronize the transactions to the global ledger
--- PASS: TestAddAndGetMessage (0.12s)
PASS
```

### FPBFT Package
FPBFT is a comprehensive simulation of the PBFT (Practical Byzantine Fault Tolerance) algorithm, designed to emulate real-world network conditions in distributed consensus scenarios. Unlike other PBFT implementations, FPBFT emphasizes the importance of varying network conditions, specifically network latency and bandwidth limit, which significantly impact the performance and fault tolerance of a distributed system. In real-world scenarios, nodes in a distributed network are dispersed across various geographical regions, each experiencing different network conditions. FPBFT integrates these parameters into the PBFT network generation, thereby setting itself apart as a full implementation of the PBFT algorithm.
#### Functionality Implemented:
> In addition to PBFT package, the FPBFT package emulated a real network by adding bandwidthLimit and latency 
> as input parameters when generating the network.

Field | Data Types | Sample Value
----  |------------| ----------
numNodes  | int        | 30-1000
data  | string     | "transactions"
clientAddr  | string     | "127.0.0.1:8888"
bandwidthLimit  | float64    | 20 (Mbps)
latency  | float64    | 350 (ms)

```go
synctime := genPBFTSynchronize(numNodes int, data string, clientAddr string, bandwidthLimit float64, latency float64)
```
> Note: by setting `bandwidthLimit` and `latency` to 0, 
> the function becomes PBFT as a special case.

#### fpbft_test.go
```go
package fpbft

import (
  "strconv"
  "testing"
)

func TestAddAndGetMessage(t *testing.T) {
  var clientAddr = "127.0.0.1:8888"
  var data = "transactions to be synchronized"
  var numNodes = 10
  sync_time := genPBFTSynchronize(numNodes, data, clientAddr, 0.01, 300)
  s := strconv.FormatFloat(sync_time, 'f', -1, 64)
  println("It takes " + s + " seconds to synchronize the transactions to the global ledger")
}

```
#### output
```text
=== RUN   TestAddAndGetMessage
initiating client...
The primary node has received a request from the client...
The request has been stored in the temporary message pool.
Broadcasting PrePrepare to other nodes...
PrePrepare broadcast completed.
It takes 2.264332033 seconds to synchronize the transactions to the global ledger
--- PASS: TestAddAndGetMessage (2.26s)
PASS
```

It will install TYPO3 into the `./myshop/` directory. Change into the directory and install TYPO3 as usual:

```bash
cd ./myshop
touch public/FIRST_INSTALL
```

Open the TYPO3 URL in your browser and follow the setup steps. Afterwards, install the Aimeos extension using:

```bash
composer req aimeos/aimeos-typo3:~23.4
```

If composer complains that one or more packages can't be installed because the required minimum stability isn't met, add this to your `composer.json`:

```json
"minimum-stability": "dev",
"prefer-stable": true,
```

If you want a more or less working installation out of the box for new installations, you can install the Bootstrap package too:

```bash
composer req bk2k/bootstrap-package
```

***Note***: Remember to create a root page and a root template, which includes the Bootstrap package templates! (See also below.)

Finally, depending on your TYPO3 version, run the following commands from your installation directory:

**For TYPO3 11:**

```bash
php ./vendor/bin/typo3 extension:setup
php ./vendor/bin/typo3 aimeos:setup --option=setup/default/demo:1
```

If you don't want to add the Aimeos demo data, you should remove `--option=setup/default/demo:1` from the Aimeos setup command.

**For TYPO3 10:**

```bash
php ./vendor/bin/typo3 extension:activate scheduler
php ./vendor/bin/typo3 extension:activate aimeos
```

If you experience any errors with the database, please check the [Database Setup](#database-setup) section below.

Please keep on reading below the "TER Extension" installation section!

### DDev

*Note:* Installation instructions for TYPO3 with `ddev` or `Colima` can be found here:
[TYPO3 with ddev or colima](https://ddev.readthedocs.io/en/latest/users/quickstart/)

### TER Extension

If you want to install Aimeos into a traditionally installed TYPO3 ("legacy installation"), the [Aimeos extension from the TER](https://typo3.org/extensions/repository/view/aimeos) is recommended. You can download and install it directly from the Extension Manager of your TYPO3 instance.

* Log into the TYPO3 backend
* Click on "Admin Tools::Extensions" in the left navigation
* Click the icon with the little plus sign left from the Aimeos list entry

![Install Aimeos TYPO3 extension](https://user-images.githubusercontent.com/213803/211545083-d0820b63-26f2-453e-877f-ecd5ec128713.jpg)

Afterwards, you have to execute the update script of the extension to create the required database structure:

* Click on "Admin Tools::Upgrade"
* Click "Run Upgrade Wizard" in the "Upgrade Wizard" tile
* Click "Execute"

![Execute update script](https://user-images.githubusercontent.com/213803/211545122-8fd94abd-78b2-47ad-ad3c-1ef1b9c052b4.jpg)

#### Aimeos Distribution

For new TYPO3 installations, there is a 1-click [Aimeos distribution](https://typo3.org/extensions/repository/view/aimeos_dist) available, too. Choose the Aimeos distribution from the list of available distributions in the Extension Manager and you will get a completely set up shop system including demo data for a quick start.

## TYPO3 Setup

Setup TYPO3 by creating a `FIRST_INSTALL` file in the `./public` directory:

```bash
touch public/FIRST_INSTALL
```

Open the URL of your installation in the browser and follow the steps in the TYPO3 setup scripts.

### Database Setup

If you use MySQL < 5.7.8, you have to use `utf8` and `utf8_unicode_ci` instead because those MySQL versions can't handle the long indexes created by `utf8mb4` (up to four bytes per character) and you will get errors like

```
1071 Specified key was too long; max key length is 767 bytes
```

To avoid that, change your database settings in your `./typo3conf/LocalConfiguration.php` to:

```php
    'DB' => [
        'Connections' => [
            'Default' => [
                'tableoptions' => [
                    'charset' => 'utf8',
                    'collate' => 'utf8_unicode_ci',
                ],
                // ...
            ],
        ],
    ],
```

### Security

Since **TYPO3 9.5.14+** implements **SameSite cookie handling** and restricts when browsers send cookies to your site. This is a problem when customers are redirected from external payment provider domain. Then, there's no session available on the confirmation page. To circumvent that problem, you need to set the configuration option `cookieSameSite` to `none` in your `./typo3conf/LocalConfiguration.php`:

```php
    'FE' => [
        'cookieSameSite' => 'none'
    ]
```

## Site Setup

TYPO3 10+ requires a site configuration which you have to add in "Site Management" > "Sites" available in the left navigation. When creating a root page (a page with a globe icon), a basic site configuration is automatically created (see below at [Go to the Import View](#go-to-the-import-view)).

## Page Setup

### Download the Aimeos Page Tree t3d file

The page setup for an Aimeos web shop is easy, if you import the example page tree for TYPO3 10/11. You can download the version you need from here:

* [23.4+ page tree](https://aimeos.org/fileadmin/download/Aimeos-pages_2023.04.t3d) and later
* [22.10 page tree](https://aimeos.org/fileadmin/download/Aimeos-pages_2022.10.t3d)
* [21.10 page tree](https://aimeos.org/fileadmin/download/Aimeos-pages_21.10.t3d)

**Note:** The Aimeos layout expects [Bootstrap](https://getbootstrap.com) providing the grid layout!

In order to upload and install the file, follow the following steps:

### Go to the Import View

**Note:** It is recommended to import the Aimeos page tree to a page that is defined as "root page". To create a root page, simply create a new page and, in the "Edit page properties", activate the "Use as Root Page" option under "Behaviour". The icon of the root page will change to a globe. This will also create a basic site configuration. Don't forget to also create a typoscript root template and include the bootstrap templates with it!

![Create a root page](https://user-images.githubusercontent.com/213803/211549273-1d3883dd-710c-4e27-8dbb-3de6e45680d7.jpg)

* In "Web::Page", right-click on the root page (the one with the globe)
* Click on "More options..."
* Click on "Import"

![Go to the import view](https://user-images.githubusercontent.com/213803/211550212-df6daa73-74cd-459e-8d25-a56c413c175d.jpg)

### Upload the page tree file

* In the page import dialog
* Select the "Upload" tab (2nd one)
* Click on the "Select" dialog
* Choose the T3D file you've downloaded
* Press the "Upload files" button

![Upload the page tree file](https://user-images.githubusercontent.com/8647429/212347778-17238e05-7494-4413-adb3-a54b2b524e05.png)

### Import the page tree

* In Import / Export view
* Select the uploaded file from the drop-down menu
* Click on the "Preview" button
* The pages that will be imported are shown below
* Click on the "Import" button that has appeared
* Confirm to import the pages

![Import the uploaded page tree file](https://user-images.githubusercontent.com/8647429/212348040-c3e10b60-5579-4d1b-becc-72548826c6db.png)

Now you have a new page "Shop" in your page tree including all required sub-pages.

### SEO-friendly URLs

TYPO3 9.5 and later can create SEO friendly URLs if you add the rules to the site config:
[https://aimeos.org/docs/latest/typo3/setup/#seo-urls](https://aimeos.org/docs/latest/typo3/setup/#seo-urls)

## License

The Aimeos TYPO3 extension is licensed under the terms of the GPL Open Source
license and is available for free.

## Links

* [Web site](https://aimeos.org/integrations/typo3-shop-extension/)
* [Documentation](https://aimeos.org/docs/TYPO3)
* [Forum](https://aimeos.org/help/typo3-extension-f16/)
* [Issue tracker](https://github.com/aimeos/aimeos-typo3/issues)
* [Source code](https://github.com/aimeos/aimeos-typo3)
