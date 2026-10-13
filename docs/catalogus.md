# Waterketen ontdekken

Gegenereerd uit de gedeelde CLI-/TUI-catalogus. Deze uitleg beschrijft de geregistreerde bronnen en bekende lagen, niet alle lagen van elke externe dienst. Lezen doet geen download of live controle.

Aanbieder → API-dienst (toegang) → laag (verzameling) → feature (kaartobject) → eigenschap (bijvoorbeeld materiaal). Een GeoPackage is een lokaal bestand, geen API.

## Vragen

### Waar liggen rioolbuizen? (`riool`)

#### Waar liggen rioolbuizen?

Bekijk de geregistreerde ligging; een publicatie bewijst niet dat iedere buis is opgenomen.

#### Rioolbuizen

Laag-ID: riool · officieel: Rotterdam-beheer:beheer_leiding

Wat is het? Lijnen met de geregistreerde ligging en eigenschappen van rioolbuizen.

Beantwoordt: Waar liggen de geregistreerde rioolbuizen?

Op de kaart: lijnen

> Wat vertelt het niet? Geen bewijs van een volledig fysiek netwerk; geen KLIC-vervanging.

In de tool: Leesbaar in Go als managed data of kaartbibliotheek; ophalen via Python.

Herkomst: GWSW-publicatie Rotterdam → GWSW netwerk

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

BOB — binnenonderkant buis — De hoogte van de binnenkant van de buisbodem ten opzichte van NAP; niet automatisch de diepte onder de straat.

NAP — Normaal Amsterdams Peil — Het landelijke referentieniveau waarmee hoogtes worden vergeleken. Het is niet het lokale straatniveau.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_leiding&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_leiding — officiële laag-ID

version=2.0.0 — versie van die standaard

### Wie levert hier drinkwater? (`drinkwater`)

#### Wie levert hier drinkwater?

Een verzorgingsgebied vertelt wie er levert, niet waar de buizen liggen.

#### Drinkwaterverzorgingsgebied

Laag-ID: drinkwater · officieel: Waterbedrijven_Nederland / FeatureServer/0

Wat is het? Gebieden die aangeven welk drinkwaterbedrijf waar levert; geen leidingen.

Beantwoordt: Wie levert hier drinkwater?

Op de kaart: vlakken

> Wat vertelt het niet? Vertelt niet waar buizen liggen. Zuid-Holland biedt een tweede levering van dezelfde soort gebieden, geen extra netwerk.

In de tool: Managed laag in Go; geen automatische omzetting naar een leidingkaart.

Herkomst: Vewin / ArcGIS-publicatie → Vewin drinkwater

Licentie: zie bron (Vewin/ArcGIS) — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Laag / collectie — Een verzameling samenhangende kaartobjecten. Eén API-dienst kan meerdere lagen aanbieden.

ArcGIS REST — FeatureServer / MapServer — Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/Waterbedrijven_Nederland/FeatureServer/0

Volledige aanvraag: https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/Waterbedrijven_Nederland/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### Waar liggen drinkwaterleidingen? (`drinkwaterleidingen`)

#### Waar liggen drinkwaterleidingen?

De gemeentelijke Evides-publicatie toont een deel van de leidinggeometrie.

#### Drinkwaterleidingen (Evides)

Laag-ID: drinkwaterleidingen · officieel: Kans_opwarming_drinkwater_Rotterdam / FeatureServer/0

Wat is het? Gepubliceerde leidinggeometrie van Evides via gemeente Rotterdam.

Beantwoordt: Waar liggen de gepubliceerde drinkwaterleidingen?

Op de kaart: lijnen

> Wat vertelt het niet? Alleen leidingen die volgens de bron niet binnen tien jaar worden vervangen. Geen volledig net; risico is een opwarmingsklasse, geen materiaal of buisdiepte.

In de tool: Go leest de lokale levering; Python haalt deze op.

Herkomst: Gemeente Rotterdam; leidingen afkomstig van Evides → Gemeentelijke Evides-publicatie

Licentie: zie bron (gemeente Rotterdam/Evides) — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

ArcGIS REST — FeatureServer / MapServer — Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

KLIC — Kabels en Leidingen Informatie Centrum — Informatie-uitwisseling via het Kadaster over kabels en leidingen, onder meer voor graafwerk.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Kans_opwarming_drinkwater_Rotterdam/FeatureServer/0

Volledige aanvraag: https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Kans_opwarming_drinkwater_Rotterdam/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### Waar liggen gemalen voor oppervlaktewater? (`oppervlaktewatergemalen`)

#### Waar liggen gemalen voor oppervlaktewater?

Deze gemeentelijke laag registreert oppervlaktewatergemalen; rioolgemalen staan apart.

#### Oppervlaktewatergemalen

Laag-ID: oppervlaktewatergemalen · officieel: Gemalen_derde_ronde / FeatureServer/0

Wat is het? Geregistreerde pompinstallaties voor oppervlaktewater, niet rioolgemalen.

Beantwoordt: Waar liggen oppervlaktewatergemalen?

Op de kaart: punten

> Wat vertelt het niet? De gemeentelijke dienst leverde bij controle 23 punten van dit type in Rotterdam, waarvan één in de wijk Delfshaven. De dienst vermeldt geen licentie of volledigheid.

In de tool: Go-kaartlaag zodra opgenomen in een lokale bibliotheeksnapshot via Python bibliotheekbouw met --waterketen.

Herkomst: Gemeente Rotterdam → Gemeente Rotterdam — oppervlaktewatergemalen

Licentie: onbekend — geen licentieveld in de dienst — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

ArcGIS REST — FeatureServer / MapServer — Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

Oppervlaktewatergemaal — Een installatie die water in het oppervlaktewatersysteem verplaatst voor peil- of waterbeheer; geen afvalwaterpomp.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Gemalen_derde_ronde/FeatureServer/0

Volledige aanvraag: https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Gemalen_derde_ronde/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### Waar wordt afvalwater gepompt of gezuiverd? (`afvalwater`)

#### Waar wordt afvalwater gepompt of gezuiverd?

Putten, gemalen en afvoergebieden vertellen verschillende delen van het verhaal. Nabijheid is geen bewezen verbinding.

#### Rioolputten

Laag-ID: putten · officieel: Rotterdam-beheer:beheer_put

Wat is het? Geregistreerde putten als kaartobjecten.

Beantwoordt: Waar liggen geregistreerde rioolputten?

Op de kaart: punten of objectgeometrie uit de levering

> Wat vertelt het niet? Een nabijgelegen buis is niet automatisch met de put verbonden; daarvoor zijn objectrelaties nodig.

In de tool: Go-kaartbibliotheek, indien deze laag is opgenomen.

Herkomst: GWSW-publicatie Rotterdam → GWSW overige Rotterdam-lagen

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_put&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_put — officiële laag-ID

version=2.0.0 — versie van die standaard

#### Gemalen / pompconstructies

Laag-ID: gemalen · officieel: Rotterdam-beheer:beheer_pomp

Wat is het? Constructies waarin afvalwater kan worden opgepompt.

Beantwoordt: Waar zijn pompconstructies geregistreerd?

Op de kaart: punten of objectgeometrie uit de levering

> Wat vertelt het niet? Deze bronlaag mengt rioolgemalen en putconstructies; een kaartobject is niet per definitie een gemaal of losse pomp.

In de tool: Go-kaartbibliotheek, indien aanwezig. RIONED heeft ook afzonderlijke collecties voor rioolgemaal en rwzi.

Herkomst: GWSW-publicatie Rotterdam → GWSW overige Rotterdam-lagen

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

RWZI — rioolwaterzuiveringsinstallatie — Een installatie die afvalwater zuivert. Een zuiveringseenheid als gebied is niet hetzelfde als een RWZI-punt.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_pomp&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_pomp — officiële laag-ID

version=2.0.0 — versie van die standaard

#### DAMO rioolgemalen

Laag-ID: damo-rioolgemaal · officieel: waterschappen-waterketen-damo:rioolgemaal

Wat is het? Geregistreerde rioolgemalen van waterschappen.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: punten

> Wat vertelt het niet? 2.815 gemeten. Gemalen zijn geen losse pomp-onderdelen; pomp staat niet in deze WFS.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO rioolgemalen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:rioolgemaal&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:rioolgemaal — officiële laag-ID

version=2.0.0 — versie van die standaard

#### DAMO zuiveringseenheden

Laag-ID: damo-zuiveringseenheid · officieel: waterschappen-waterketen-damo:zuiveringseenheid

Wat is het? Gebieden die als zuiveringseenheid zijn geregistreerd.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: vlakken

> Wat vertelt het niet? Geen RWZI-punten en geen bewezen route van een individuele buis.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:zuiveringseenheid&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:zuiveringseenheid — officiële laag-ID

version=2.0.0 — versie van die standaard

#### DAMO RWZI's

Laag-ID: damo-rwzi · officieel: waterschappen-waterketen-damo:rwzi

Wat is het? Rioolwaterzuiveringsinstallaties als puntobject.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? 317 gemeten: één punt per installatie, zonder capaciteit of prestatie.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO RWZI's

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:rwzi&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:rwzi — officiële laag-ID

version=2.0.0 — versie van die standaard

#### DAMO lozingspunten

Laag-ID: damo-lozingspunt · officieel: waterschappen-waterketen-damo:lozingspunt

Wat is het? Lozingen op oppervlaktewater, inclusief overstorten.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? 3.236 gemeten. Een lozingspunt bewijst geen overstortwerking.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO lozingspunten

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:lozingspunt&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:lozingspunt — officiële laag-ID

version=2.0.0 — versie van die standaard

#### RIONED rioolgemalen

Laag-ID: rioned-rioolgemaal · officieel: waterschap_rioolgemaal

Wat is het? Rioolgemalen als eigen GWSW-collectie.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Alleen liveness: PDOK geeft geen totaaltelling.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED rioolgemalen

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_rioolgemaal/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_rioolgemaal/items: objecten uit deze collectie (gegevenslaag).

#### RIONED RWZI's

Laag-ID: rioned-rwzi · officieel: waterschap_rwzi

Wat is het? RWZI's als eigen GWSW-collectie.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Alleen liveness; procesgegevens staan niet in deze collectie.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED RWZI's

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_rwzi/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_rwzi/items: objecten uit deze collectie (gegevenslaag).

#### Begrippen bij dit antwoord

RWZI — rioolwaterzuiveringsinstallatie — Een installatie die afvalwater zuivert. Een zuiveringseenheid als gebied is niet hetzelfde als een RWZI-punt.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

### Welk water, welke kering en welk peilgebied liggen bij dit gebied? (`watercontext`)

#### Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Oppervlaktewater, keringen en peilgebieden beschrijven de omgeving van het netwerk; ze zijn geen registratie van de buizen zelf.

#### Oppervlaktewaterlichaam (IMWA)

Laag-ID: hwh-oppervlaktewaterlichaam · officieel: waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam

Wat is het? Landelijke laag van oppervlaktewaterlichamen volgens IMWA.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: vlakken

> Wat vertelt het niet? 1.430.861 gemeten: beheerssemantiek, geen topografie. Geen debiet of waterstand.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Oppervlaktewateren IMWA

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam — officiële laag-ID

version=2.0.0 — versie van die standaard

#### Waterkering (IMWA)

Laag-ID: hwh-waterkering · officieel: waterschappen-keringen-imwa:waterkering

Wat is het? Waterkeringen zoals dijken, dammen en kades.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: lijnen en vlakken

> Wat vertelt het niet? 28.396 gemeten. Registratie is geen actuele toetsstatus.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Keringen IMWA

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-keringen-imwa:waterkering&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-keringen-imwa:waterkering — officiële laag-ID

version=2.0.0 — versie van die standaard

#### Peilgebied (IMWA)

Laag-ID: hwh-peilgebied · officieel: waterschappen-waterbeheergebieden-imwa:peilgebied

Wat is het? Gebieden met één afgesproken waterpeil.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: vlakken

> Wat vertelt het niet? 37.418 gemeten. Grenzen vallen niet samen met waterschapsgrenzen.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Waterbeheergebieden IMWA (peilgebied)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterbeheergebieden-imwa:peilgebied&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterbeheergebieden-imwa:peilgebied — officiële laag-ID

version=2.0.0 — versie van die standaard

#### Kunstwerken IMWA — gemalen

Laag-ID: hwh-kunstwerken-gemaal · officieel: waterschappen-kunstwerken-imwa:gemaal

Wat is het? Door waterschappen beheerde gemalen in het watersysteem.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: punten

> Wat vertelt het niet? 8.565 gemeten. Een gemaal hier is niet automatisch een rioolgemaal.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Kunstwerken IMWA (gemaal)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-kunstwerken-imwa:gemaal&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-kunstwerken-imwa:gemaal — officiële laag-ID

version=2.0.0 — versie van die standaard

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

### Hoe hoog of diep liggen de buizen? (`diepte`)

#### Hoe hoog of diep liggen de buizen?

BOB is buisbodemhoogte ten opzichte van NAP. Voor diepte onder een straat is ook het maaiveld nodig.

#### Rioolbuizen

Laag-ID: riool · officieel: Rotterdam-beheer:beheer_leiding

Wat is het? Lijnen met de geregistreerde ligging en eigenschappen van rioolbuizen.

Beantwoordt: Waar liggen de geregistreerde rioolbuizen?

Op de kaart: lijnen

> Wat vertelt het niet? Geen bewijs van een volledig fysiek netwerk; geen KLIC-vervanging.

In de tool: Leesbaar in Go als managed data of kaartbibliotheek; ophalen via Python.

Herkomst: GWSW-publicatie Rotterdam → GWSW netwerk

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

BOB — binnenonderkant buis — De hoogte van de binnenkant van de buisbodem ten opzichte van NAP; niet automatisch de diepte onder de straat.

NAP — Normaal Amsterdams Peil — Het landelijke referentieniveau waarmee hoogtes worden vergeleken. Het is niet het lokale straatniveau.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_leiding&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_leiding — officiële laag-ID

version=2.0.0 — versie van die standaard

#### Terreinhoogte (AHN DTM)

Laag-ID: hoogte · officieel: dtm_05m

Wat is het? Een raster met terreinhoogtes om met buisbodemhoogtes te vergelijken.

Beantwoordt: Hoe ver ligt de buisbodem onder het straatmaaiveld?

Op de kaart: raster; geen buizen

> Wat vertelt het niet? BOB alleen is geen diepte. Wateroppervlak en straatniveau moeten apart worden behandeld.

In de tool: Alleen de Python-opdracht diepte; deze doet een expliciete download.

Herkomst: AHN via PDOK → AHN DTM (WCS)

Licentie: onbekend — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

AHN — Actueel Hoogtebestand Nederland — Hoogtegegevens van Nederland. Het DTM beschrijft het terrein; het DSM bevat ook objecten boven het terrein.

WCS — Web Coverage Service — Een standaard om rastergegevens (een rooster met waarden) op te vragen.

NAP — Normaal Amsterdams Peil — Het landelijke referentieniveau waarmee hoogtes worden vergeleken. Het is niet het lokale straatniveau.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/rws/ahn/wcs/v1_0

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

#### Begrippen bij dit antwoord

Maaiveld — Het oppervlak van de grond. Voor straatdiepte moet je de buisbodem met het straatniveau vergelijken.

### Wat ligt er boven en rond het netwerk? (`omgeving`)

#### Wat ligt er boven en rond het netwerk?

Straten, water en wijkgrenzen geven context; ze zijn geen registratie van ondergrondse leidingen.

#### Straten

Laag-ID: straten · officieel: OSM highway

Wat is het? Straatgeometrie als bovengrondse oriëntatie.

Beantwoordt: Wat ligt er boven en rond het netwerk?

Op de kaart: lijnen

> Wat vertelt het niet? Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.

In de tool: Go-kaartbibliotheek, indien geïmporteerd via Python.

Herkomst: OpenStreetMap-bijdragers → OpenStreetMap (lokale PBF)

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GeoPackage (GPKG) — Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

#### Oppervlaktewater

Laag-ID: water · officieel: OSM natural=water / landuse=reservoir

Wat is het? Vlakken voor water zoals singels, havens en reservoirs.

Beantwoordt: Wat ligt er boven en rond het netwerk?

Op de kaart: vlakken

> Wat vertelt het niet? Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.

In de tool: Go-kaartbibliotheek, indien geïmporteerd via Python.

Herkomst: OpenStreetMap-bijdragers → OpenStreetMap (lokale PBF)

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GeoPackage (GPKG) — Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

#### Gemeente, wijken en buurten

Laag-ID: gebieden · officieel: wijkenbuurten:gemeenten / wijkenbuurten:wijken / wijkenbuurten:buurten

Wat is het? Officiële CBS-gebiedsgrenzen voor ruimtelijke selectie.

Beantwoordt: Welke wijk of buurt onderzoek ik?

Op de kaart: vlakken

> Wat vertelt het niet? Een bbox is niet de exacte gebiedsgrens. Een zoekvak-telling bewijst geen wijkdekking.

In de tool: Gebieden kiezen uit de lokale kaartbibliotheek; bouwen via Python.

Herkomst: CBS via PDOK → CBS wijk- en buurtgrenzen (2025)

Licentie: onbekend — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

CRS — coördinatenreferentiesysteem — De afspraak die bepaalt hoe coördinaten op de aarde worden geplaatst.

Voorbeeldlaag in deze aanvraag: wijkenbuurten:wijken

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/cbs/wijkenbuurten/2025/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/cbs/wijkenbuurten/2025/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=wijkenbuurten:wijken&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=wijkenbuurten:wijken — officiële laag-ID

version=2.0.0 — versie van die standaard

#### Begrippen bij dit antwoord

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

## Bronnen

### GWSW netwerk (`gws_netwerk`)

#### GWSW netwerk

Bron-ID: gws_netwerk · aanbieder: GWSW-publicatie Rotterdam

Geregistreerde rioolbuizen en hun eigenschappen.

> Beperking: Geen bewijs van fysieke volledigheid; publieke data is geen KLIC-vervanging.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Rioolbuizen (riool) — Lijnen met de geregistreerde ligging en eigenschappen van rioolbuizen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_leiding&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_leiding — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO waterketen (`damo_leiding`)

#### DAMO waterketen

Bron-ID: damo_leiding · aanbieder: Het Waterschapshuis / waterschappen via PDOK

Waterketenobjecten volgens DAMO, zoals leidingen, putten en gemalen.

> Beperking: Kan overlappen met GWSW. Niet iedere beschreven objectsoort is via WFS beschikbaar.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: DAMO leidingen (damo-leiding) — Geregistreerde waterketenleidingen van waterschappen volgens DAMO.

Laag: DAMO putten (damo-put) — Waterschapsputten als afzonderlijke levering.

Laag: DAMO rioleringsgebieden (damo-rioleringsgebied) — Gebieden die de afvoer van afvalwater beschrijven.

Laag: DAMO zuiveringseenheden (damo-zuiveringseenheid) — Gebieden die als zuiveringseenheid zijn geregistreerd.

Laag: DAMO mantelbuizen (damo-mantelbuis) — Beschermbuis rond een leiding, bijvoorbeeld bij een kruising.

Laag: DAMO ontluchters (damo-ontluchter) — Ontluchting op een transportleiding.

Laag: DAMO leidingsegmenten (damo-leidingsegment) — Segmenten van waterketenleidingen, los van het leidingobject.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:leiding&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:leiding — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO rioolgemalen (`damo_rioolgemaal`)

#### DAMO rioolgemalen

Bron-ID: damo_rioolgemaal · aanbieder: Het Waterschapshuis / waterschappen via PDOK

Rioolgemalen die waterschappen volgens DAMO registreren.

> Beperking: Losse pomp-onderdelen staan niet in deze WFS. Kan hetzelfde fysieke object zijn als een GWSW-gemaal.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: DAMO rioolgemalen (damo-rioolgemaal) — Geregistreerde rioolgemalen van waterschappen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:rioolgemaal&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:rioolgemaal — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO RWZI's (`damo_rwzi`)

#### DAMO RWZI's

Bron-ID: damo_rwzi · aanbieder: Het Waterschapshuis / waterschappen via PDOK

Rioolwaterzuiveringsinstallaties als puntobject.

> Beperking: Eén punt per installatie; capaciteit of prestatie staat er niet in. Niet optellen bij de CBS-statistiek.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: DAMO RWZI's (damo-rwzi) — Rioolwaterzuiveringsinstallaties als puntobject.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:rwzi&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:rwzi — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO lozingspunten (`damo_lozingspunt`)

#### DAMO lozingspunten

Bron-ID: damo_lozingspunt · aanbieder: Het Waterschapshuis / waterschappen via PDOK

Punten waar afvalwater op het oppervlaktewater loost, inclusief overstorten.

> Beperking: Een lozingspunt is geen bewezen overstortwerking; de bron beschrijft de lozing, niet het gedrag.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: DAMO lozingspunten (damo-lozingspunt) — Lozingen op oppervlaktewater, inclusief overstorten.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:lozingspunt&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:lozingspunt — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO overnamepunten (`damo_overnamepunt`)

#### DAMO overnamepunten

Bron-ID: damo_overnamepunt · aanbieder: Het Waterschapshuis / waterschappen via PDOK

Punten waar gemeentelijk riool overgaat naar het waterschapssysteem.

> Beperking: Markeert de ketenoverdracht, niet de volledige route van het water daarna.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: DAMO overnamepunten (damo-overnamepunt) — Overdracht van gemeentelijk riool naar het waterschapssysteem.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:overnamepunt&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:overnamepunt — officiële laag-ID

version=2.0.0 — versie van die standaard

### Vewin drinkwater (`vewin`)

#### Vewin drinkwater

Bron-ID: vewin · aanbieder: Vewin / ArcGIS-publicatie

Verzorgingsgebieden van drinkwaterbedrijven.

> Beperking: Gebieden, geen leidingen. Gebruiksvoorwaarden zijn niet expliciet vastgesteld.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: zie bron (Vewin/ArcGIS) — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Drinkwaterverzorgingsgebied (drinkwater) — Gebieden die aangeven welk drinkwaterbedrijf waar levert; geen leidingen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/Waterbedrijven_Nederland/FeatureServer/0

Volledige aanvraag: https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/Waterbedrijven_Nederland/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### RIONED Waterketen GWSW (OGC) (`rioned_gwsw`)

#### RIONED Waterketen GWSW (OGC)

Bron-ID: rioned_gwsw · aanbieder: Stichting RIONED via PDOK

Waterketenobjecten volgens GWSW, toegankelijk als afzonderlijke collecties.

> Beperking: De kleine probe geeft geen totaalaantal; landelijke levering zegt niets over wijkdekking.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: RIONED waterschapsleidingen (rioned-leiding) — Een afzonderlijke collectie van waterschapsleidingen volgens GWSW.

Laag: RIONED putten (rioned-put) — Putten als eigen GWSW-collectie.

Laag: RIONED lozingspunten (rioned-lozingspunt) — Lozingen als eigen GWSW-collectie.

Laag: RIONED overnamepunten (rioned-overnamepunt) — Overnamepunten als eigen GWSW-collectie.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_leiding/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_leiding/items: objecten uit deze collectie (gegevenslaag).

### RIONED rioolgemalen (`rioned_rioolgemaal`)

#### RIONED rioolgemalen

Bron-ID: rioned_rioolgemaal · aanbieder: Stichting RIONED via PDOK

Rioolgemalen als eigen GWSW-collectie.

> Beperking: Lost wat de GWSW beheer_pomp-laag vermengt: die telt gemalen en putconstructies bij elkaar op.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: RIONED rioolgemalen (rioned-rioolgemaal) — Rioolgemalen als eigen GWSW-collectie.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_rioolgemaal/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_rioolgemaal/items: objecten uit deze collectie (gegevenslaag).

### RIONED RWZI's (`rioned_rwzi`)

#### RIONED RWZI's

Bron-ID: rioned_rwzi · aanbieder: Stichting RIONED via PDOK

RWZI's als eigen GWSW-collectie.

> Beperking: Alleen de installatie; procesgegevens staan in de CBS-tabellen, niet hier.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: RIONED RWZI's (rioned-rwzi) — RWZI's als eigen GWSW-collectie.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_rwzi/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_rwzi/items: objecten uit deze collectie (gegevenslaag).

### RIONED IBA's (`rioned_iba`)

#### RIONED IBA's

Bron-ID: rioned_iba · aanbieder: Stichting RIONED via PDOK

Individuele behandeling afvalwater: kleinschalige systemen buiten het riool.

> Beperking: Een IBA is geen rioolaansluiting; dekking hangt af van wat de bronhouder aanlevert.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: RIONED IBA's (rioned-iba) — Kleinschalige afvalwaterbehandeling buiten het riool.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_iba/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_iba/items: objecten uit deze collectie (gegevenslaag).

### Zuid-Holland drinkwater (WFS) (`zh_drinkwater`)

#### Zuid-Holland drinkwater (WFS)

Bron-ID: zh_drinkwater · aanbieder: Provincie Zuid-Holland

Drinkwaterbedrijven, winnings- en beschermingsinformatie en een afzonderlijke transportlevering.

> Beperking: Verzorgingsgebieden kunnen dezelfde soort gegevens als Vewin leveren; dit is geen extra leidingnet. Licenties gelden per dataset.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: Publiek domein (PDM 1.0) — publiek-domeinmarkering; geen auteursrechtelijke beperkingen bekend

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: ZH drinkwaterverzorgingsgebied (zh-drinkwater-gebied) — Een tweede levering van drinkwaterverzorgingsgebieden, naast Vewin.

Laag: ZH drinkwaterproductieketen (zh-drinkwater-productieketen) — Productie- en winningslocaties voor drinkwater.

Laag: ZH drinkwaterzoekgebieden (zh-drinkwater-zoekgebieden) — Zoekruimte voor drinkwaterwinning.

Laag: ZH grondwaterbeschermingsgebieden (zh-grondwaterbescherming) — Gebieden voor bescherming van grondwater.

Laag: ZH Delfland leidingsegmenten (zh-delfland-transport) — Een afzonderlijke publicatie van transportleidingsegmenten.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.zuid-holland.nl/geoserver/water/wfs

Volledige aanvraag: https://geodata.zuid-holland.nl/geoserver/water/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=water:DRINKWATER_BEDRIJVEN&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=water:DRINKWATER_BEDRIJVEN — officiële laag-ID

version=2.0.0 — versie van die standaard

### Waterschappen Kunstwerken IMWA (gemaal) (`hwh_kunstwerken_imwa`)

#### Waterschappen Kunstwerken IMWA (gemaal)

Bron-ID: hwh_kunstwerken_imwa · aanbieder: Het Waterschapshuis via PDOK

Door waterschappen beheerde kunstwerken volgens IMWA: gemalen, stuwen, sluizen, duikers.

> Beperking: De probe telt alleen de gemaal-laag; de dienst bevat 18 objectsoorten. Registratie zegt niets over actuele bediening of toestand.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Kunstwerken IMWA — gemalen (hwh-kunstwerken-gemaal) — Door waterschappen beheerde gemalen in het watersysteem.

Laag: Kunstwerken IMWA — stuwen (hwh-kunstwerken-stuw) — Stuwen die het waterpeil in een watergang regelen.

Laag: Kunstwerken IMWA — sluizen (hwh-kunstwerken-sluis) — Sluizen in het waterbeheer en de scheepvaart.

Laag: Kunstwerken IMWA — duikers (hwh-kunstwerken-duiker) — Duikers: watergangen die onder een weg of kering door kruisen.

Laag: Kunstwerken IMWA — vispassages (hwh-kunstwerken-vispassage) — Voorzieningen die vismigratie langs een kunstwerk mogelijk maken.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-kunstwerken-imwa:gemaal&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-kunstwerken-imwa:gemaal — officiële laag-ID

version=2.0.0 — versie van die standaard

### Waterschappen Oppervlaktewateren IMWA (`hwh_oppervlaktewateren_imwa`)

#### Waterschappen Oppervlaktewateren IMWA

Bron-ID: hwh_oppervlaktewateren_imwa · aanbieder: Het Waterschapshuis via PDOK

Landelijke laag van oppervlaktewaterlichamen en oevers volgens IMWA.

> Beperking: Beheerssemantiek, geen topografie: de BGT-waterdelen zijn geen duplicaat. Geen debiet of waterstand.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Oppervlaktewaterlichaam (IMWA) (hwh-oppervlaktewaterlichaam) — Landelijke laag van oppervlaktewaterlichamen volgens IMWA.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam — officiële laag-ID

version=2.0.0 — versie van die standaard

### Waterschappen Keringen IMWA (`hwh_keringen_imwa`)

#### Waterschappen Keringen IMWA

Bron-ID: hwh_keringen_imwa · aanbieder: Het Waterschapshuis via PDOK

Waterkeringen en waterkeringsecties van de waterschappen.

> Beperking: Een geregistreerde kering is geen actuele toetsstatus; gebruik de legger voor juridische uitspraken.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Waterkering (IMWA) (hwh-waterkering) — Waterkeringen zoals dijken, dammen en kades.

Laag: Waterkeringsectie (IMWA) (hwh-waterkeringsectie) — Secties waarin een waterkering is opgedeeld.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-keringen-imwa:waterkering&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-keringen-imwa:waterkering — officiële laag-ID

version=2.0.0 — versie van die standaard

### Waterschappen Waterbeheergebieden IMWA (peilgebied) (`hwh_waterbeheergebieden_imwa`)

#### Waterschappen Waterbeheergebieden IMWA (peilgebied)

Bron-ID: hwh_waterbeheergebieden_imwa · aanbieder: Het Waterschapshuis via PDOK

Peilgebieden, peilbesluitgebieden en afvoer-/aanvoergebieden.

> Beperking: De probe telt de peilgebiedlaag. Grenzen verschillen van bestuurlijke waterschapsgrenzen.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Peilgebied (IMWA) (hwh-peilgebied) — Gebieden met één afgesproken waterpeil.

Laag: Peilbesluitgebied (IMWA) (hwh-peilbesluitgebied) — Het deel waarvoor formeel een peilbesluit geldt.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterbeheergebieden-imwa:peilgebied&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterbeheergebieden-imwa:peilgebied — officiële laag-ID

version=2.0.0 — versie van die standaard

### RWS regiogebieden (ArcGIS) (`rws_regiogebieden`)

#### RWS regiogebieden (ArcGIS)

Bron-ID: rws_regiogebieden · aanbieder: Rijkswaterstaat

Grenzen van diensten en districten van Rijkswaterstaat.

> Beperking: Beheergebieden zijn geen drinkwaternet. De dienst vermeldt geen expliciete licentie.

Toegang in de tool: Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus.

Licentie: onbekend — geen licentieveld in de dienst — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: RWS regiogebieden (rws-regiogebieden) — Beheergrenzen van Rijkswaterstaatdiensten en districten.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geo.rijkswaterstaat.nl/arcgis/rest/services/GDR/regiogebieden_rijkswaterstaat/MapServer/0

Volledige aanvraag: https://geo.rijkswaterstaat.nl/arcgis/rest/services/GDR/regiogebieden_rijkswaterstaat/MapServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### Waterschappen Hydrografie (INSPIRE) (`hwh_hydrografie`)

#### Waterschappen Hydrografie (INSPIRE)

Bron-ID: hwh_hydrografie · aanbieder: Het Waterschapshuis via PDOK

Een geharmoniseerd netwerk van waterlopen volgens INSPIRE.

> Beperking: CC BY-NC-ND verbiedt commercieel gebruik en afgeleide werken; bewust niet aangesloten.

Toegang in de tool: Geen live broncontrole aangesloten: bewust niet bedraad — licentie eerst beslissen

Licentie: CC BY-NC-ND 4.0 (geen commercieel gebruik, geen afgeleiden) — geen commercieel gebruik en geen afgeleide werken

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/hwh/waterschappen-hydrografie/ogc/v1

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### RIONED landelijk stedelijk water (`rioned_national`)

#### RIONED landelijk stedelijk water

Bron-ID: rioned_national · aanbieder: Stichting RIONED via PDOK

Een landelijke GWSW-levering als één GeoPackage-bestand.

> Beperking: Een bulkbestand van circa 2,99 GB; nooit automatisch downloaden. Niet verwarren met DAMO.

Toegang in de tool: Geen live broncontrole aangesloten: ~2,99 GB bulk — alleen expliciet, nooit impliciet

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/rioned/beheer-stedelijk-watersystemen-gwsw/atom/downloads/stedelijkwater.gpkg

Bestandsdownload: https://service.pdok.nl/rioned/beheer-stedelijk-watersystemen-gwsw/atom/downloads/stedelijkwater.gpkg

> Een groot GeoPackage-bestand; alleen expliciet downloaden. Geen live status of telling gecontroleerd.

### KLIC-afgeleid Rotterdam (ArcGIS) (`klic_rotterdam`)

#### KLIC-afgeleid Rotterdam (ArcGIS)

Bron-ID: klic_rotterdam · aanbieder: Persoonlijk/Hogeschool-account op ArcGIS

Een KLIC-afgeleide rioolpublicatie met andere herkomst dan de primaire bron.

> Beperking: Fallback, niet de primaire gemeentelijke registratie. Rechten en herkomst eerst beoordelen.

Toegang in de tool: Geen live broncontrole aangesloten: provenance: persoonlijk/Hogeschool-account — fallback, niet primair

Licentie: zie bron (persoonlijk/Hogeschool-account) — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services9.arcgis.com/nqW2A97fCpeCbw6Z/arcgis/rest/services/Gemeente_Rotterdam_Rioolbuizen/FeatureServer/0

Volledige aanvraag: https://services9.arcgis.com/nqW2A97fCpeCbw6Z/arcgis/rest/services/Gemeente_Rotterdam_Rioolbuizen/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### GWSW overige Rotterdam-lagen (`gws_aanvullend`)

#### GWSW overige Rotterdam-lagen

Bron-ID: gws_aanvullend · aanbieder: GWSW-publicatie Rotterdam

Aanvullende Rotterdamse lagen, waaronder putten en pompconstructies.

> Beperking: De pompconstructielaag kan zowel rioolgemalen als putconstructies bevatten.

Toegang in de tool: Geen live broncontrole aangesloten: 27 lagen gedetecteerd; alleen beheer/netwerk/waterschap bedraad

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Rioolputten (putten) — Geregistreerde putten als kaartobjecten.

Laag: Gemalen / pompconstructies (gemalen) — Constructies waarin afvalwater kan worden opgepompt.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_put&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_put — officiële laag-ID

version=2.0.0 — versie van die standaard

### RWS drinkwater (NWP) (`rws_nwp_drinkwater`)

#### RWS drinkwater (NWP)

Bron-ID: rws_nwp_drinkwater · aanbieder: Rijkswaterstaat

Drinkwaterinnamepunten en beschermingsinformatie.

> Beperking: Geen volledig leidingnet; licentie eerst verifiëren.

Toegang in de tool: Geen live broncontrole aangesloten: innamepunten, punten en beschermingszones — licentie eerst verifiëren

Licentie: onbekend — geen licentieveld in de dienst — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geo.rijkswaterstaat.nl/arcgis/rest/services/GDR/nwp_drinkwater/MapServer

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### Waterschappen Waterschapsgrenzen IMSO (alleen WMS) (`hwh_waterschapsgrenzen_imso`)

#### Waterschappen Waterschapsgrenzen IMSO (alleen WMS)

Bron-ID: hwh_waterschapsgrenzen_imso · aanbieder: Het Waterschapshuis via PDOK

Bestuurlijke waterschapsgrenzen volgens IMSO.

> Beperking: Alleen als WMS beschikbaar: wel te tekenen, niet als objecten op te vragen.

Toegang in de tool: Geen live broncontrole aangesloten: alleen WMS, geen WFS: geen objectvraag; attributie loopt via GWSW/DAMO

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterschapsgrenzen-imso/wms/v2_0

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### Gemeentelijke Evides-publicatie (`evides`)

#### Gemeentelijke Evides-publicatie

Bron-ID: evides · aanbieder: Gemeente Rotterdam; leidingen afkomstig van Evides

Lijngeometrie van een deel van de drinkwaterleidingen, met opwarmingsrisico.

> Beperking: Volgens de bron alleen leidingen die niet binnen tien jaar worden vervangen; geen volledige netwerk- of buisattributenregistratie.

Toegang in de tool: Ophalen via Python; geen afzonderlijke live probe in Go.

Licentie: zie bron (gemeente Rotterdam/Evides) — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Drinkwaterleidingen (Evides) (drinkwaterleidingen) — Gepubliceerde leidinggeometrie van Evides via gemeente Rotterdam.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Kans_opwarming_drinkwater_Rotterdam/FeatureServer/0

Volledige aanvraag: https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Kans_opwarming_drinkwater_Rotterdam/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### OpenStreetMap (lokale PBF) (`osm`)

#### OpenStreetMap (lokale PBF)

Bron-ID: osm · aanbieder: OpenStreetMap-bijdragers

Bovengrondse context uit een lokaal geïmporteerd provinciaal bestand.

> Beperking: Geen volledige officiële objectregistratie of ondergrondse leidingkaart.

Toegang in de tool: Python importeert een lokale PBF; Go leest de bestaande snapshot. Geen API-aanvraag.

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Straten (straten) — Straatgeometrie als bovengrondse oriëntatie.

Laag: Oppervlaktewater (water) — Vlakken voor water zoals singels, havens en reservoirs.

Laag: Groen (groen) — Geselecteerde groene gebieden zoals parken en bossen.

Laag: Gebouwen (gebouwen) — Gebouwcontouren als bovengrondse context.

Laag: Spoor (spoor) — Spoor-, tram- en metrolijnen uit OpenStreetMap.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### AHN DTM (WCS) (`ahn`)

#### AHN DTM (WCS)

Bron-ID: ahn · aanbieder: AHN via PDOK

Terreinhoogtes als raster, voor vergelijking met de buisbodem.

> Beperking: Boven open water is een hoogtesample niet automatisch het straatmaaiveld. Geen licentie in dit register vastgesteld.

Toegang in de tool: De Python-opdracht diepte downloadt AHN expliciet; geen Go-kaartlaag.

Licentie: onbekend — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Terreinhoogte (AHN DTM) (hoogte) — Een raster met terreinhoogtes om met buisbodemhoogtes te vergelijken.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/rws/ahn/wcs/v1_0

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### CBS wijk- en buurtgrenzen (2025) (`cbs`)

#### CBS wijk- en buurtgrenzen (2025)

Bron-ID: cbs · aanbieder: CBS via PDOK

Gemeente-, wijk- en buurtgrenzen met officiële gebiedscodes.

> Beperking: 2025 is het standaard bouwjaar in Python, niet een claim over het jaar van de lokale snapshot. Licentie hier niet vastgesteld.

Toegang in de tool: Python haalt de grenzen op bij bibliotheekbouw; Go leest de snapshot.

Licentie: onbekend — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Gemeente, wijken en buurten (gebieden) — Officiële CBS-gebiedsgrenzen voor ruimtelijke selectie.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/cbs/wijkenbuurten/2025/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/cbs/wijkenbuurten/2025/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=wijkenbuurten:wijken&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=wijkenbuurten:wijken — officiële laag-ID

version=2.0.0 — versie van die standaard

### Gemeente Rotterdam — oppervlaktewatergemalen (`rotterdam_oppervlaktewatergemalen`)

#### Gemeente Rotterdam — oppervlaktewatergemalen

Bron-ID: rotterdam_oppervlaktewatergemalen · aanbieder: Gemeente Rotterdam

Punten van oppervlaktewatergemalen, apart van rioolgemalen.

> Beperking: De dienst specificeert geen licentie of volledigheid; gebruik de bronmetadata en beschouw de telling niet als compleetheidsbewijs.

Toegang in de tool: Wordt als aparte laag opgenomen bij Python bibliotheekbouw met --waterketen; Go leest daarna de nieuwe snapshot.

Licentie: onbekend — geen licentieveld in de dienst — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

> Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.

Laag: Oppervlaktewatergemalen (oppervlaktewatergemalen) — Geregistreerde pompinstallaties voor oppervlaktewater, niet rioolgemalen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Gemalen_derde_ronde/FeatureServer/0

Volledige aanvraag: https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Gemalen_derde_ronde/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

## Lagen

### Rioolbuizen (`riool`)

#### Rioolbuizen

Laag-ID: riool · officieel: Rotterdam-beheer:beheer_leiding

Wat is het? Lijnen met de geregistreerde ligging en eigenschappen van rioolbuizen.

Beantwoordt: Waar liggen de geregistreerde rioolbuizen?

Op de kaart: lijnen

> Wat vertelt het niet? Geen bewijs van een volledig fysiek netwerk; geen KLIC-vervanging.

In de tool: Leesbaar in Go als managed data of kaartbibliotheek; ophalen via Python.

Herkomst: GWSW-publicatie Rotterdam → GWSW netwerk

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

BOB — binnenonderkant buis — De hoogte van de binnenkant van de buisbodem ten opzichte van NAP; niet automatisch de diepte onder de straat.

NAP — Normaal Amsterdams Peil — Het landelijke referentieniveau waarmee hoogtes worden vergeleken. Het is niet het lokale straatniveau.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_leiding&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_leiding — officiële laag-ID

version=2.0.0 — versie van die standaard

### Drinkwaterverzorgingsgebied (`drinkwater`)

#### Drinkwaterverzorgingsgebied

Laag-ID: drinkwater · officieel: Waterbedrijven_Nederland / FeatureServer/0

Wat is het? Gebieden die aangeven welk drinkwaterbedrijf waar levert; geen leidingen.

Beantwoordt: Wie levert hier drinkwater?

Op de kaart: vlakken

> Wat vertelt het niet? Vertelt niet waar buizen liggen. Zuid-Holland biedt een tweede levering van dezelfde soort gebieden, geen extra netwerk.

In de tool: Managed laag in Go; geen automatische omzetting naar een leidingkaart.

Herkomst: Vewin / ArcGIS-publicatie → Vewin drinkwater

Licentie: zie bron (Vewin/ArcGIS) — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Laag / collectie — Een verzameling samenhangende kaartobjecten. Eén API-dienst kan meerdere lagen aanbieden.

ArcGIS REST — FeatureServer / MapServer — Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/Waterbedrijven_Nederland/FeatureServer/0

Volledige aanvraag: https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/Waterbedrijven_Nederland/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### Drinkwaterleidingen (Evides) (`drinkwaterleidingen`)

#### Drinkwaterleidingen (Evides)

Laag-ID: drinkwaterleidingen · officieel: Kans_opwarming_drinkwater_Rotterdam / FeatureServer/0

Wat is het? Gepubliceerde leidinggeometrie van Evides via gemeente Rotterdam.

Beantwoordt: Waar liggen de gepubliceerde drinkwaterleidingen?

Op de kaart: lijnen

> Wat vertelt het niet? Alleen leidingen die volgens de bron niet binnen tien jaar worden vervangen. Geen volledig net; risico is een opwarmingsklasse, geen materiaal of buisdiepte.

In de tool: Go leest de lokale levering; Python haalt deze op.

Herkomst: Gemeente Rotterdam; leidingen afkomstig van Evides → Gemeentelijke Evides-publicatie

Licentie: zie bron (gemeente Rotterdam/Evides) — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

ArcGIS REST — FeatureServer / MapServer — Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

KLIC — Kabels en Leidingen Informatie Centrum — Informatie-uitwisseling via het Kadaster over kabels en leidingen, onder meer voor graafwerk.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Kans_opwarming_drinkwater_Rotterdam/FeatureServer/0

Volledige aanvraag: https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Kans_opwarming_drinkwater_Rotterdam/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### Rioolputten (`putten`)

#### Rioolputten

Laag-ID: putten · officieel: Rotterdam-beheer:beheer_put

Wat is het? Geregistreerde putten als kaartobjecten.

Beantwoordt: Waar liggen geregistreerde rioolputten?

Op de kaart: punten of objectgeometrie uit de levering

> Wat vertelt het niet? Een nabijgelegen buis is niet automatisch met de put verbonden; daarvoor zijn objectrelaties nodig.

In de tool: Go-kaartbibliotheek, indien deze laag is opgenomen.

Herkomst: GWSW-publicatie Rotterdam → GWSW overige Rotterdam-lagen

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_put&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_put — officiële laag-ID

version=2.0.0 — versie van die standaard

### Gemalen / pompconstructies (`gemalen`)

#### Gemalen / pompconstructies

Laag-ID: gemalen · officieel: Rotterdam-beheer:beheer_pomp

Wat is het? Constructies waarin afvalwater kan worden opgepompt.

Beantwoordt: Waar zijn pompconstructies geregistreerd?

Op de kaart: punten of objectgeometrie uit de levering

> Wat vertelt het niet? Deze bronlaag mengt rioolgemalen en putconstructies; een kaartobject is niet per definitie een gemaal of losse pomp.

In de tool: Go-kaartbibliotheek, indien aanwezig. RIONED heeft ook afzonderlijke collecties voor rioolgemaal en rwzi.

Herkomst: GWSW-publicatie Rotterdam → GWSW overige Rotterdam-lagen

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

RWZI — rioolwaterzuiveringsinstallatie — Een installatie die afvalwater zuivert. Een zuiveringseenheid als gebied is niet hetzelfde als een RWZI-punt.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.gwsw.nl/geoserver/ows

Volledige aanvraag: https://geodata.gwsw.nl/geoserver/ows?service=WFS&version=2.0.0&request=GetFeature&typeNames=Rotterdam-beheer:beheer_pomp&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=Rotterdam-beheer:beheer_pomp — officiële laag-ID

version=2.0.0 — versie van die standaard

### Terreinhoogte (AHN DTM) (`hoogte`)

#### Terreinhoogte (AHN DTM)

Laag-ID: hoogte · officieel: dtm_05m

Wat is het? Een raster met terreinhoogtes om met buisbodemhoogtes te vergelijken.

Beantwoordt: Hoe ver ligt de buisbodem onder het straatmaaiveld?

Op de kaart: raster; geen buizen

> Wat vertelt het niet? BOB alleen is geen diepte. Wateroppervlak en straatniveau moeten apart worden behandeld.

In de tool: Alleen de Python-opdracht diepte; deze doet een expliciete download.

Herkomst: AHN via PDOK → AHN DTM (WCS)

Licentie: onbekend — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

AHN — Actueel Hoogtebestand Nederland — Hoogtegegevens van Nederland. Het DTM beschrijft het terrein; het DSM bevat ook objecten boven het terrein.

WCS — Web Coverage Service — Een standaard om rastergegevens (een rooster met waarden) op te vragen.

NAP — Normaal Amsterdams Peil — Het landelijke referentieniveau waarmee hoogtes worden vergeleken. Het is niet het lokale straatniveau.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/rws/ahn/wcs/v1_0

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### Gemeente, wijken en buurten (`gebieden`)

#### Gemeente, wijken en buurten

Laag-ID: gebieden · officieel: wijkenbuurten:gemeenten / wijkenbuurten:wijken / wijkenbuurten:buurten

Wat is het? Officiële CBS-gebiedsgrenzen voor ruimtelijke selectie.

Beantwoordt: Welke wijk of buurt onderzoek ik?

Op de kaart: vlakken

> Wat vertelt het niet? Een bbox is niet de exacte gebiedsgrens. Een zoekvak-telling bewijst geen wijkdekking.

In de tool: Gebieden kiezen uit de lokale kaartbibliotheek; bouwen via Python.

Herkomst: CBS via PDOK → CBS wijk- en buurtgrenzen (2025)

Licentie: onbekend — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

CRS — coördinatenreferentiesysteem — De afspraak die bepaalt hoe coördinaten op de aarde worden geplaatst.

Voorbeeldlaag in deze aanvraag: wijkenbuurten:wijken

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/cbs/wijkenbuurten/2025/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/cbs/wijkenbuurten/2025/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=wijkenbuurten:wijken&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=wijkenbuurten:wijken — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO leidingen (`damo-leiding`)

#### DAMO leidingen

Laag-ID: damo-leiding · officieel: waterschappen-waterketen-damo:leiding

Wat is het? Geregistreerde waterketenleidingen van waterschappen volgens DAMO.

Beantwoordt: Welke waterschapsleidingen zijn gepubliceerd?

Op de kaart: lijnen

> Wat vertelt het niet? Kan dezelfde fysieke objecten bevatten als GWSW; tellingen niet optellen als nieuw netwerk.

In de tool: Broncontrole in Go; geen afzonderlijke DAMO-leidinglaag in de huidige Go-kaart.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:leiding&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:leiding — officiële laag-ID

version=2.0.0 — versie van die standaard

### RIONED waterschapsleidingen (`rioned-leiding`)

#### RIONED waterschapsleidingen

Laag-ID: rioned-leiding · officieel: waterschap_leiding

Wat is het? Een afzonderlijke collectie van waterschapsleidingen volgens GWSW.

Beantwoordt: Welke leidingen publiceert RIONED via PDOK?

Op de kaart: lijnen

> Wat vertelt het niet? Een antwoord met limit=1 is geen totaaltelling en bewijst geen wijkdekking.

In de tool: Broncontrole in Go; niet automatisch als extra netwerk op de kaart geladen.

Herkomst: Stichting RIONED via PDOK → RIONED Waterketen GWSW (OGC)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

GWSW — Gegevenswoordenboek Stedelijk Water — Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

OGC API Features — Een standaard voor het opvragen van kaartobjecten via collecties en webadressen. Een collectie is een gegevenslaag.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_leiding/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_leiding/items: objecten uit deze collectie (gegevenslaag).

### ZH drinkwaterverzorgingsgebied (`zh-drinkwater-gebied`)

#### ZH drinkwaterverzorgingsgebied

Laag-ID: zh-drinkwater-gebied · officieel: water:DRINKWATER_BEDRIJVEN

Wat is het? Een tweede levering van drinkwaterverzorgingsgebieden, naast Vewin.

Beantwoordt: Wie levert hier drinkwater?

Op de kaart: vlakken

> Wat vertelt het niet? Geen leidingen en geen nieuw netwerk. Deze levering kan dienen als kruiscontrole op Vewin.

In de tool: Broncontrole in Go; geen extra leidingkaart.

Herkomst: Provincie Zuid-Holland → Zuid-Holland drinkwater (WFS)

Licentie: Publiek domein (PDM 1.0) — publiek-domeinmarkering; geen auteursrechtelijke beperkingen bekend

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

Laag / collectie — Een verzameling samenhangende kaartobjecten. Eén API-dienst kan meerdere lagen aanbieden.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.zuid-holland.nl/geoserver/water/wfs

Volledige aanvraag: https://geodata.zuid-holland.nl/geoserver/water/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=water:DRINKWATER_BEDRIJVEN&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=water:DRINKWATER_BEDRIJVEN — officiële laag-ID

version=2.0.0 — versie van die standaard

### RWS regiogebieden (`rws-regiogebieden`)

#### RWS regiogebieden

Laag-ID: rws-regiogebieden · officieel: regiogebieden_rijkswaterstaat / MapServer/0

Wat is het? Beheergrenzen van Rijkswaterstaatdiensten en districten.

Beantwoordt: Welke Rijkswaterstaatregio beheert dit gebied?

Op de kaart: vlakken

> Wat vertelt het niet? Geen CBS-wijkgrenzen, verzorgingsgebied of leidingnet. De dienst publiceert geen licentieveld.

In de tool: Broncontrole in Go; geen lokale snapshotlaag zolang deze niet expliciet is opgenomen.

Herkomst: Rijkswaterstaat → RWS regiogebieden (ArcGIS)

Licentie: onbekend — geen licentieveld in de dienst — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

ArcGIS REST — FeatureServer / MapServer — Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

Laag / collectie — Een verzameling samenhangende kaartobjecten. Eén API-dienst kan meerdere lagen aanbieden.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geo.rijkswaterstaat.nl/arcgis/rest/services/GDR/regiogebieden_rijkswaterstaat/MapServer/0

Volledige aanvraag: https://geo.rijkswaterstaat.nl/arcgis/rest/services/GDR/regiogebieden_rijkswaterstaat/MapServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

### Straten (`straten`)

#### Straten

Laag-ID: straten · officieel: OSM highway

Wat is het? Straatgeometrie als bovengrondse oriëntatie.

Beantwoordt: Wat ligt er boven en rond het netwerk?

Op de kaart: lijnen

> Wat vertelt het niet? Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.

In de tool: Go-kaartbibliotheek, indien geïmporteerd via Python.

Herkomst: OpenStreetMap-bijdragers → OpenStreetMap (lokale PBF)

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GeoPackage (GPKG) — Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### Oppervlaktewater (`water`)

#### Oppervlaktewater

Laag-ID: water · officieel: OSM natural=water / landuse=reservoir

Wat is het? Vlakken voor water zoals singels, havens en reservoirs.

Beantwoordt: Wat ligt er boven en rond het netwerk?

Op de kaart: vlakken

> Wat vertelt het niet? Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.

In de tool: Go-kaartbibliotheek, indien geïmporteerd via Python.

Herkomst: OpenStreetMap-bijdragers → OpenStreetMap (lokale PBF)

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GeoPackage (GPKG) — Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### Groen (`groen`)

#### Groen

Laag-ID: groen · officieel: OSM natural / landuse / leisure-selectie

Wat is het? Geselecteerde groene gebieden zoals parken en bossen.

Beantwoordt: Wat ligt er boven en rond het netwerk?

Op de kaart: vlakken

> Wat vertelt het niet? Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.

In de tool: Go-kaartbibliotheek, indien geïmporteerd via Python.

Herkomst: OpenStreetMap-bijdragers → OpenStreetMap (lokale PBF)

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GeoPackage (GPKG) — Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### Gebouwen (`gebouwen`)

#### Gebouwen

Laag-ID: gebouwen · officieel: OSM building

Wat is het? Gebouwcontouren als bovengrondse context.

Beantwoordt: Wat ligt er boven en rond het netwerk?

Op de kaart: vlakken

> Wat vertelt het niet? Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.

In de tool: Go-kaartbibliotheek, indien geïmporteerd via Python.

Herkomst: OpenStreetMap-bijdragers → OpenStreetMap (lokale PBF)

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GeoPackage (GPKG) — Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### Spoor (`spoor`)

#### Spoor

Laag-ID: spoor · officieel: OSM railway

Wat is het? Spoor-, tram- en metrolijnen uit OpenStreetMap.

Beantwoordt: Wat ligt er boven en rond het netwerk?

Op de kaart: lijnen

> Wat vertelt het niet? Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.

In de tool: Go-kaartbibliotheek, indien geïmporteerd via Python.

Herkomst: OpenStreetMap-bijdragers → OpenStreetMap (lokale PBF)

Licentie: ODbL — naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

GeoPackage (GPKG) — Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig):

Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.

### DAMO putten (`damo-put`)

#### DAMO putten

Laag-ID: damo-put · officieel: waterschappen-waterketen-damo:put

Wat is het? Waterschapsputten als afzonderlijke levering.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: punten

> Wat vertelt het niet? Mogelijke overlap met GWSW; niet automatisch samenvoegen.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:put&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:put — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO rioolgemalen (`damo-rioolgemaal`)

#### DAMO rioolgemalen

Laag-ID: damo-rioolgemaal · officieel: waterschappen-waterketen-damo:rioolgemaal

Wat is het? Geregistreerde rioolgemalen van waterschappen.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: punten

> Wat vertelt het niet? 2.815 gemeten. Gemalen zijn geen losse pomp-onderdelen; pomp staat niet in deze WFS.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO rioolgemalen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:rioolgemaal&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:rioolgemaal — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO rioleringsgebieden (`damo-rioleringsgebied`)

#### DAMO rioleringsgebieden

Laag-ID: damo-rioleringsgebied · officieel: waterschappen-waterketen-damo:rioleringsgebied

Wat is het? Gebieden die de afvoer van afvalwater beschrijven.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: vlakken

> Wat vertelt het niet? Geen buizennetwerk; een afvoergebied is een vlak.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:rioleringsgebied&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:rioleringsgebied — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO zuiveringseenheden (`damo-zuiveringseenheid`)

#### DAMO zuiveringseenheden

Laag-ID: damo-zuiveringseenheid · officieel: waterschappen-waterketen-damo:zuiveringseenheid

Wat is het? Gebieden die als zuiveringseenheid zijn geregistreerd.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: vlakken

> Wat vertelt het niet? Geen RWZI-punten en geen bewezen route van een individuele buis.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:zuiveringseenheid&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:zuiveringseenheid — officiële laag-ID

version=2.0.0 — versie van die standaard

### ZH drinkwaterproductieketen (`zh-drinkwater-productieketen`)

#### ZH drinkwaterproductieketen

Laag-ID: zh-drinkwater-productieketen · officieel: water:DRINKWATER_PRODUCTIEKETEN

Wat is het? Productie- en winningslocaties voor drinkwater.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: punten

> Wat vertelt het niet? Geen drinkwaterleidingen; PDM geldt voor deze beschreven dataset, niet automatisch voor alle dienstlagen.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Provincie Zuid-Holland → Zuid-Holland drinkwater (WFS)

Licentie: Publiek domein (PDM 1.0) — publiek-domeinmarkering; geen auteursrechtelijke beperkingen bekend

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.zuid-holland.nl/geoserver/water/wfs

Volledige aanvraag: https://geodata.zuid-holland.nl/geoserver/water/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=water:DRINKWATER_PRODUCTIEKETEN&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=water:DRINKWATER_PRODUCTIEKETEN — officiële laag-ID

version=2.0.0 — versie van die standaard

### ZH drinkwaterzoekgebieden (`zh-drinkwater-zoekgebieden`)

#### ZH drinkwaterzoekgebieden

Laag-ID: zh-drinkwater-zoekgebieden · officieel: water:DRINKWATER_ZOEKGEBIEDEN

Wat is het? Zoekruimte voor drinkwaterwinning.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: vlakken

> Wat vertelt het niet? Geen verzorgingsgebied of bewijs dat hier al water wordt gewonnen.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Provincie Zuid-Holland → Zuid-Holland drinkwater (WFS)

Licentie: Publiek domein (PDM 1.0) — publiek-domeinmarkering; geen auteursrechtelijke beperkingen bekend

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.zuid-holland.nl/geoserver/water/wfs

Volledige aanvraag: https://geodata.zuid-holland.nl/geoserver/water/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=water:DRINKWATER_ZOEKGEBIEDEN&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=water:DRINKWATER_ZOEKGEBIEDEN — officiële laag-ID

version=2.0.0 — versie van die standaard

### ZH grondwaterbeschermingsgebieden (`zh-grondwaterbescherming`)

#### ZH grondwaterbeschermingsgebieden

Laag-ID: zh-grondwaterbescherming · officieel: water:BESCHERMINGSGEBIED_GRONDWATER

Wat is het? Gebieden voor bescherming van grondwater.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: vlakken

> Wat vertelt het niet? Geen leidingnet; de grens toont bescherming, niet de ligging van buizen.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Provincie Zuid-Holland → Zuid-Holland drinkwater (WFS)

Licentie: Publiek domein (PDM 1.0) — publiek-domeinmarkering; geen auteursrechtelijke beperkingen bekend

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.zuid-holland.nl/geoserver/water/wfs

Volledige aanvraag: https://geodata.zuid-holland.nl/geoserver/water/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=water:BESCHERMINGSGEBIED_GRONDWATER&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=water:BESCHERMINGSGEBIED_GRONDWATER — officiële laag-ID

version=2.0.0 — versie van die standaard

### ZH Delfland leidingsegmenten (`zh-delfland-transport`)

#### ZH Delfland leidingsegmenten

Laag-ID: zh-delfland-transport · officieel: water:HHDELFLAND_LEIDINGSEGMENT

Wat is het? Een afzonderlijke publicatie van transportleidingsegmenten.

Beantwoordt: Welke aanvullende waterketengegevens zijn gepubliceerd?

Op de kaart: lijnen

> Wat vertelt het niet? Spaarzame levering met mogelijke overlap; geen nieuw drinkwaternet. Bbox-matches kunnen buiten de wijk liggen.

In de tool: Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.

Herkomst: Provincie Zuid-Holland → Zuid-Holland drinkwater (WFS)

Licentie: Publiek domein (PDM 1.0) — publiek-domeinmarkering; geen auteursrechtelijke beperkingen bekend

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Bbox — bounding box — Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://geodata.zuid-holland.nl/geoserver/water/wfs

Volledige aanvraag: https://geodata.zuid-holland.nl/geoserver/water/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=water:HHDELFLAND_LEIDINGSEGMENT&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=water:HHDELFLAND_LEIDINGSEGMENT — officiële laag-ID

version=2.0.0 — versie van die standaard

### Kunstwerken IMWA — gemalen (`hwh-kunstwerken-gemaal`)

#### Kunstwerken IMWA — gemalen

Laag-ID: hwh-kunstwerken-gemaal · officieel: waterschappen-kunstwerken-imwa:gemaal

Wat is het? Door waterschappen beheerde gemalen in het watersysteem.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: punten

> Wat vertelt het niet? 8.565 gemeten. Een gemaal hier is niet automatisch een rioolgemaal.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Kunstwerken IMWA (gemaal)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-kunstwerken-imwa:gemaal&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-kunstwerken-imwa:gemaal — officiële laag-ID

version=2.0.0 — versie van die standaard

### Kunstwerken IMWA — stuwen (`hwh-kunstwerken-stuw`)

#### Kunstwerken IMWA — stuwen

Laag-ID: hwh-kunstwerken-stuw · officieel: waterschappen-kunstwerken-imwa:stuw

Wat is het? Stuwen die het waterpeil in een watergang regelen.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: punten

> Wat vertelt het niet? 53.142 gemeten: de grootste groep. Een stuw regelt peil, geen afvalwatertransport.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Kunstwerken IMWA (gemaal)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-kunstwerken-imwa:stuw&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-kunstwerken-imwa:stuw — officiële laag-ID

version=2.0.0 — versie van die standaard

### Kunstwerken IMWA — sluizen (`hwh-kunstwerken-sluis`)

#### Kunstwerken IMWA — sluizen

Laag-ID: hwh-kunstwerken-sluis · officieel: waterschappen-kunstwerken-imwa:sluis

Wat is het? Sluizen in het waterbeheer en de scheepvaart.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: punten

> Wat vertelt het niet? 889 gemeten. Geen zuivering of gemaal; een eigen objecttype.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Kunstwerken IMWA (gemaal)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-kunstwerken-imwa:sluis&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-kunstwerken-imwa:sluis — officiële laag-ID

version=2.0.0 — versie van die standaard

### Kunstwerken IMWA — duikers (`hwh-kunstwerken-duiker`)

#### Kunstwerken IMWA — duikers

Laag-ID: hwh-kunstwerken-duiker · officieel: waterschappen-kunstwerken-imwa:duiker

Wat is het? Duikers: watergangen die onder een weg of kering door kruisen.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: punten

> Wat vertelt het niet? Een duiker is geen rioolbuis, ook al ligt hij ondergronds.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Kunstwerken IMWA (gemaal)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-kunstwerken-imwa:duiker&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-kunstwerken-imwa:duiker — officiële laag-ID

version=2.0.0 — versie van die standaard

### Kunstwerken IMWA — vispassages (`hwh-kunstwerken-vispassage`)

#### Kunstwerken IMWA — vispassages

Laag-ID: hwh-kunstwerken-vispassage · officieel: waterschappen-kunstwerken-imwa:vispassage

Wat is het? Voorzieningen die vismigratie langs een kunstwerk mogelijk maken.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: punten

> Wat vertelt het niet? Ecologische inrichting; geen waterketenobject.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Kunstwerken IMWA (gemaal)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-kunstwerken-imwa:vispassage&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-kunstwerken-imwa:vispassage — officiële laag-ID

version=2.0.0 — versie van die standaard

### Oppervlaktewaterlichaam (IMWA) (`hwh-oppervlaktewaterlichaam`)

#### Oppervlaktewaterlichaam (IMWA)

Laag-ID: hwh-oppervlaktewaterlichaam · officieel: waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam

Wat is het? Landelijke laag van oppervlaktewaterlichamen volgens IMWA.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: vlakken

> Wat vertelt het niet? 1.430.861 gemeten: beheerssemantiek, geen topografie. Geen debiet of waterstand.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Oppervlaktewateren IMWA

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam — officiële laag-ID

version=2.0.0 — versie van die standaard

### Waterkering (IMWA) (`hwh-waterkering`)

#### Waterkering (IMWA)

Laag-ID: hwh-waterkering · officieel: waterschappen-keringen-imwa:waterkering

Wat is het? Waterkeringen zoals dijken, dammen en kades.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: lijnen en vlakken

> Wat vertelt het niet? 28.396 gemeten. Registratie is geen actuele toetsstatus.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Keringen IMWA

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-keringen-imwa:waterkering&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-keringen-imwa:waterkering — officiële laag-ID

version=2.0.0 — versie van die standaard

### Waterkeringsectie (IMWA) (`hwh-waterkeringsectie`)

#### Waterkeringsectie (IMWA)

Laag-ID: hwh-waterkeringsectie · officieel: waterschappen-keringen-imwa:waterkeringsectie

Wat is het? Secties waarin een waterkering is opgedeeld.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: lijnen

> Wat vertelt het niet? 21.825 gemeten: een sectie hoort bij één kering, dus niet optellen als extra kering.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Keringen IMWA

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-keringen-imwa:waterkeringsectie&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-keringen-imwa:waterkeringsectie — officiële laag-ID

version=2.0.0 — versie van die standaard

### Peilgebied (IMWA) (`hwh-peilgebied`)

#### Peilgebied (IMWA)

Laag-ID: hwh-peilgebied · officieel: waterschappen-waterbeheergebieden-imwa:peilgebied

Wat is het? Gebieden met één afgesproken waterpeil.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: vlakken

> Wat vertelt het niet? 37.418 gemeten. Grenzen vallen niet samen met waterschapsgrenzen.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Waterbeheergebieden IMWA (peilgebied)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterbeheergebieden-imwa:peilgebied&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterbeheergebieden-imwa:peilgebied — officiële laag-ID

version=2.0.0 — versie van die standaard

### Peilbesluitgebied (IMWA) (`hwh-peilbesluitgebied`)

#### Peilbesluitgebied (IMWA)

Laag-ID: hwh-peilbesluitgebied · officieel: waterschappen-waterbeheergebieden-imwa:peilbesluitgebied

Wat is het? Het deel waarvoor formeel een peilbesluit geldt.

Beantwoordt: Welk water, welke kering en welk peilgebied liggen bij dit gebied?

Op de kaart: vlakken

> Wat vertelt het niet? 3.937 gemeten: het besluitdeel, geen groter peilgebied.

In de tool: Broncontrole in Go; geen automatische kaartlaag of analyse.

Herkomst: Het Waterschapshuis via PDOK → Waterschappen Waterbeheergebieden IMWA (peilgebied)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

IMWA — Informatiemodel Water — Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Kering — waterkering — Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Peilgebied — Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

WFS — Web Feature Service — Een standaard om kaartobjecten en hun eigenschappen op te vragen.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterbeheergebieden-imwa:peilbesluitgebied&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterbeheergebieden-imwa:peilbesluitgebied — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO RWZI's (`damo-rwzi`)

#### DAMO RWZI's

Laag-ID: damo-rwzi · officieel: waterschappen-waterketen-damo:rwzi

Wat is het? Rioolwaterzuiveringsinstallaties als puntobject.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? 317 gemeten: één punt per installatie, zonder capaciteit of prestatie.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO RWZI's

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:rwzi&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:rwzi — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO lozingspunten (`damo-lozingspunt`)

#### DAMO lozingspunten

Laag-ID: damo-lozingspunt · officieel: waterschappen-waterketen-damo:lozingspunt

Wat is het? Lozingen op oppervlaktewater, inclusief overstorten.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? 3.236 gemeten. Een lozingspunt bewijst geen overstortwerking.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO lozingspunten

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:lozingspunt&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:lozingspunt — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO overnamepunten (`damo-overnamepunt`)

#### DAMO overnamepunten

Laag-ID: damo-overnamepunt · officieel: waterschappen-waterketen-damo:overnamepunt

Wat is het? Overdracht van gemeentelijk riool naar het waterschapssysteem.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? 2.024 gemeten: het koppelpunt, niet de route daarna.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO overnamepunten

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:overnamepunt&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:overnamepunt — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO mantelbuizen (`damo-mantelbuis`)

#### DAMO mantelbuizen

Laag-ID: damo-mantelbuis · officieel: waterschappen-waterketen-damo:mantelbuis

Wat is het? Beschermbuis rond een leiding, bijvoorbeeld bij een kruising.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: lijnen

> Wat vertelt het niet? 3.603 gemeten. Een mantelbuis is geen transportleiding.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:mantelbuis&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:mantelbuis — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO ontluchters (`damo-ontluchter`)

#### DAMO ontluchters

Laag-ID: damo-ontluchter · officieel: waterschappen-waterketen-damo:ontluchter

Wat is het? Ontluchting op een transportleiding.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? 4.257 gemeten. Onderdeel van het transportsysteem, geen rioolput.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:ontluchter&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:ontluchter — officiële laag-ID

version=2.0.0 — versie van die standaard

### DAMO leidingsegmenten (`damo-leidingsegment`)

#### DAMO leidingsegmenten

Laag-ID: damo-leidingsegment · officieel: waterschappen-waterketen-damo:leidingsegment

Wat is het? Segmenten van waterketenleidingen, los van het leidingobject.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: lijnen

> Wat vertelt het niet? 28.221 gemeten. Segment en leiding beschrijven deels dezelfde buis; niet optellen.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Het Waterschapshuis / waterschappen via PDOK → DAMO waterketen

Licentie: CC BY-SA 4.0 — naamsvermelding en gelijk delen bij afgeleide werken

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0

Volledige aanvraag: https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0?service=WFS&version=2.0.0&request=GetFeature&typeNames=waterschappen-waterketen-damo:leidingsegment&resultType=hits

Antwoord: XML met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

request=GetFeature — soort vraag aan de dienst

resultType=hits — hits telt alleen, zonder de objecten te downloaden

service=WFS — standaard waarmee we de dienst aanspreken

typeNames=waterschappen-waterketen-damo:leidingsegment — officiële laag-ID

version=2.0.0 — versie van die standaard

### RIONED rioolgemalen (`rioned-rioolgemaal`)

#### RIONED rioolgemalen

Laag-ID: rioned-rioolgemaal · officieel: waterschap_rioolgemaal

Wat is het? Rioolgemalen als eigen GWSW-collectie.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Alleen liveness: PDOK geeft geen totaaltelling.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED rioolgemalen

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_rioolgemaal/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_rioolgemaal/items: objecten uit deze collectie (gegevenslaag).

### RIONED RWZI's (`rioned-rwzi`)

#### RIONED RWZI's

Laag-ID: rioned-rwzi · officieel: waterschap_rwzi

Wat is het? RWZI's als eigen GWSW-collectie.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Alleen liveness; procesgegevens staan niet in deze collectie.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED RWZI's

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_rwzi/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_rwzi/items: objecten uit deze collectie (gegevenslaag).

### RIONED IBA's (`rioned-iba`)

#### RIONED IBA's

Laag-ID: rioned-iba · officieel: waterschap_iba

Wat is het? Kleinschalige afvalwaterbehandeling buiten het riool.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Dekking hangt af van wat de bronhouder aanlevert.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED IBA's

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_iba/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_iba/items: objecten uit deze collectie (gegevenslaag).

### RIONED putten (`rioned-put`)

#### RIONED putten

Laag-ID: rioned-put · officieel: waterschap_put

Wat is het? Putten als eigen GWSW-collectie.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Alleen liveness; nabijheid van een buis bewijst geen verbinding.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED Waterketen GWSW (OGC)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_put/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_put/items: objecten uit deze collectie (gegevenslaag).

### RIONED lozingspunten (`rioned-lozingspunt`)

#### RIONED lozingspunten

Laag-ID: rioned-lozingspunt · officieel: waterschap_lozingspunt

Wat is het? Lozingen als eigen GWSW-collectie.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Losse objecten; geen bewijs van overstortgedrag.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED Waterketen GWSW (OGC)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_lozingspunt/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_lozingspunt/items: objecten uit deze collectie (gegevenslaag).

### RIONED overnamepunten (`rioned-overnamepunt`)

#### RIONED overnamepunten

Laag-ID: rioned-overnamepunt · officieel: waterschap_overnamepunt

Wat is het? Overnamepunten als eigen GWSW-collectie.

Beantwoordt: Welke objectsoorten publiceren de waterschappen en RIONED?

Op de kaart: punten

> Wat vertelt het niet? Koppelpunt in de keten, geen volledige route.

In de tool: Broncontrole in Go; niet automatisch op de kaart of in analyses.

Herkomst: Stichting RIONED via PDOK → RIONED Waterketen GWSW (OGC)

Licentie: CC0 — vrij hergebruik; bewaar de herkomst

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

DAMO — Datamodel van de waterschappen — Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Stichting RIONED — Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Overstort — Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

IBA — individuele behandeling afvalwater — Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0

Volledige aanvraag: https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0/collections/waterschap_overnamepunt/items?limit=1&f=json

Antwoord: JSON met maximaal één kaartobject; totaalaantal onbekend

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

limit=1 — maximum aantal objecten in dit antwoord, niet het totaalaantal

/collections/waterschap_overnamepunt/items: objecten uit deze collectie (gegevenslaag).

### Oppervlaktewatergemalen (`oppervlaktewatergemalen`)

#### Oppervlaktewatergemalen

Laag-ID: oppervlaktewatergemalen · officieel: Gemalen_derde_ronde / FeatureServer/0

Wat is het? Geregistreerde pompinstallaties voor oppervlaktewater, niet rioolgemalen.

Beantwoordt: Waar liggen oppervlaktewatergemalen?

Op de kaart: punten

> Wat vertelt het niet? De gemeentelijke dienst leverde bij controle 23 punten van dit type in Rotterdam, waarvan één in de wijk Delfshaven. De dienst vermeldt geen licentie of volledigheid.

In de tool: Go-kaartlaag zodra opgenomen in een lokale bibliotheeksnapshot via Python bibliotheekbouw met --waterketen.

Herkomst: Gemeente Rotterdam → Gemeente Rotterdam — oppervlaktewatergemalen

Licentie: onbekend — geen licentieveld in de dienst — gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron

#### Lokaal aanwezig / actualiteit

Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.

#### Begrippen bij dit antwoord

Feature — kaartobject — Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

ArcGIS REST — FeatureServer / MapServer — Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

Oppervlaktewatergemaal — Een installatie die water in het oppervlaktewatersysteem verplaatst voor peil- of waterbeheer; geen afvalwaterpomp.

#### Technische verdieping

Basisadres (parameters zijn nog nodig): https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Gemalen_derde_ronde/FeatureServer/0

Volledige aanvraag: https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Gemalen_derde_ronde/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json

Antwoord: JSON met een objecttelling; geen geometrie

> Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.

f=json — formaat van het antwoord

returnCountOnly=true — true vraagt alleen een telling

where=1=1 — 1=1 selecteert alle objecten

## Begrippen

### GWSW — Gegevenswoordenboek Stedelijk Water (`gwsw`)

#### GWSW — Gegevenswoordenboek Stedelijk Water

Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.

Voorbeeld: Een buis kan een materiaal en begin- en eindpunt hebben volgens GWSW.

### DAMO — Datamodel van de waterschappen (`damo`)

#### DAMO — Datamodel van de waterschappen

Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.

Voorbeeld: Een DAMO-rioolgemaal en een GWSW-rioolgemaal kunnen hetzelfde fysieke object beschrijven.

### Stichting RIONED (`rioned`)

#### Stichting RIONED

Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.

Voorbeeld: RIONED-gegevens kunnen via PDOK worden gepubliceerd.

### PDOK — Publieke Dienstverlening Op de Kaart (`pdok`)

#### PDOK — Publieke Dienstverlening Op de Kaart

Een voorziening die geografische gegevens van verschillende aanbieders toegankelijk maakt.

Voorbeeld: Een waterschap kan gegevens leveren die via een PDOK-dienst beschikbaar zijn.

### WFS — Web Feature Service (`wfs`)

#### WFS — Web Feature Service

Een standaard om kaartobjecten en hun eigenschappen op te vragen.

Voorbeeld: GetFeature met resultType=hits vraagt alleen het aantal objecten in een laag.

### OGC API Features (`ogc`)

#### OGC API Features

Een standaard voor het opvragen van kaartobjecten via collecties en webadressen. Een collectie is een gegevenslaag.

Voorbeeld: /collections/waterschap_leiding/items bevat objecten uit die collectie.

### ArcGIS REST — FeatureServer / MapServer (`arcgis`)

#### ArcGIS REST — FeatureServer / MapServer

Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.

Voorbeeld: Een /query-aanvraag met returnCountOnly=true geeft een telling, geen buisvormen.

### Feature — kaartobject (`feature`)

#### Feature — kaartobject

Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.

Voorbeeld: Een rioolbuis is een lijn; het materiaal is een eigenschap van die lijn.

### Laag / collectie (`layer`)

#### Laag / collectie

Een verzameling samenhangende kaartobjecten. Eén API-dienst kan meerdere lagen aanbieden.

Voorbeeld: Een laag met verzorgingsgebieden bevat vlakken, geen drinkwaterleidingen.

### API-endpoint (`endpoint`)

#### API-endpoint

Een webadres waar een programma een bepaalde vraag kan stellen. Vaak zijn ook parameters nodig.

Voorbeeld: Een kale WFS-URL kan een fout geven; een volledige GetFeature-aanvraag noemt ook de laag.

### BOB — binnenonderkant buis (`bob`)

#### BOB — binnenonderkant buis

De hoogte van de binnenkant van de buisbodem ten opzichte van NAP; niet automatisch de diepte onder de straat.

Voorbeeld: Begin- en eindhoogte horen bij verschillende buisuiteinden. In de lokale rioollevering behandelt de tool 0,00 als onbekend.

### NAP — Normaal Amsterdams Peil (`nap`)

#### NAP — Normaal Amsterdams Peil

Het landelijke referentieniveau waarmee hoogtes worden vergeleken. Het is niet het lokale straatniveau.

Voorbeeld: Een positieve BOB kan onder een hoger gelegen straat liggen; het teken alleen bewijst geen fout.

### Maaiveld (`maaiveld`)

#### Maaiveld

Het oppervlak van de grond. Voor straatdiepte moet je de buisbodem met het straatniveau vergelijken.

Voorbeeld: AHN-waarden boven open water kunnen het wateroppervlak voorstellen, niet de straat.

### AHN — Actueel Hoogtebestand Nederland (`ahn`)

#### AHN — Actueel Hoogtebestand Nederland

Hoogtegegevens van Nederland. Het DTM beschrijft het terrein; het DSM bevat ook objecten boven het terrein.

Voorbeeld: Voor maaiveld gebruikt de Python-tool het DTM, niet een gebouwhoogte uit het DSM.

### WCS — Web Coverage Service (`wcs`)

#### WCS — Web Coverage Service

Een standaard om rastergegevens (een rooster met waarden) op te vragen.

Voorbeeld: Een AHN-aanvraag kan een GeoTIFF met hoogtes opleveren, geen losse rioolbuizen.

### CRS — coördinatenreferentiesysteem (`crs`)

#### CRS — coördinatenreferentiesysteem

De afspraak die bepaalt hoe coördinaten op de aarde worden geplaatst.

Voorbeeld: RD New (EPSG:28992) gebruikt meters; WGS 84 (EPSG:4326) gebruikt geografische graden.

### Bbox — bounding box (`bbox`)

#### Bbox — bounding box

Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.

Voorbeeld: De hoeken van een wijkvak kunnen stukken van buurgemeenten of buurwijken bevatten.

### GeoPackage (GPKG) (`gpkg`)

#### GeoPackage (GPKG)

Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.

Voorbeeld: De CLI en TUI lezen dezelfde beheerde GeoPackages; het bestand is geen API.

### RWZI — rioolwaterzuiveringsinstallatie (`rwzi`)

#### RWZI — rioolwaterzuiveringsinstallatie

Een installatie die afvalwater zuivert. Een zuiveringseenheid als gebied is niet hetzelfde als een RWZI-punt.

Voorbeeld: Een vlak dat het aangesloten gebied beschrijft, markeert niet de locatie van de zuiveringsinstallatie.

### Oppervlaktewatergemaal (`oppervlaktewatergemaal`)

#### Oppervlaktewatergemaal

Een installatie die water in het oppervlaktewatersysteem verplaatst voor peil- of waterbeheer; geen afvalwaterpomp.

Voorbeeld: De gemeentelijke kaartlaag typeert oppervlaktewatergemalen apart van rioolgemalen.

### KLIC — Kabels en Leidingen Informatie Centrum (`klic`)

#### KLIC — Kabels en Leidingen Informatie Centrum

Informatie-uitwisseling via het Kadaster over kabels en leidingen, onder meer voor graafwerk.

Voorbeeld: Deze openbare kaartcatalogus vervangt geen vereiste KLIC-melding.

### Snapshot — vastgelegde levering (`snapshot`)

#### Snapshot — vastgelegde levering

Een bewaarde versie van gegevens op een bepaald moment. De ophaaldatum hoeft niet de meet- of brondatum te zijn.

Voorbeeld: Een recent opgehaald bestand kan oude metingen bevatten.

### IMWA — Informatiemodel Water (`imwa`)

#### IMWA — Informatiemodel Water

Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.

Voorbeeld: Een waterkering en een peilgebied zijn IMWA-objecttypen.

### Kering — waterkering (`kering`)

#### Kering — waterkering

Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.

Voorbeeld: Een geregistreerde kering is geen actuele toetsstatus.

### Peilgebied (`peilgebied`)

#### Peilgebied

Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.

Voorbeeld: Het peilgebied vertelt welk peil hoort bij de sloot naast een rioolbuis.

### Overstort (`overstort`)

#### Overstort

Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.

Voorbeeld: Een DAMO-lozingspunt kan een overstort zijn, maar de bron beschrijft het gedrag niet.

### IBA — individuele behandeling afvalwater (`iba`)

#### IBA — individuele behandeling afvalwater

Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.

Voorbeeld: Een IBA is geen rioolaansluiting.
