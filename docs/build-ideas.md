# Build ideas

[Project overview](../README.md) · [Generator](../gen/README.md)

A running backlog of generated builds. The target is **80,000–130,000 cells** at the
default density (2, span ≈0.794): large enough to read as a detailed model and to
give a 15–17 second assembly, small enough to load and settle quickly.

## What the first five builds teach

| Build | Cells | Notes |
| --- | ---: | --- |
| Arena | 151,483 | Heaviest; long elliptical arcades read well, but solid seating bowls cost many hidden cells |
| Castle | 138,426 | Strong silhouette from towers and colored cone roofs; mostly one sand-stone color |
| Manhattan | 98,511 | Spread over a long island; reads at a distance thanks to varied building colors |
| Cruise Ship | 66,663 | Clean white mass with color accents (lifeboats, funnel); smaller than the sweet spot |
| Medieval Tower | 34,374 | Elegant but small; most cells are a hollow mound and one tower |

Guidelines drawn from them:

- Hollow shells with 1–1.5 unit walls stretch the budget; solids should be reserved for
  thin or small parts. The renderer drops fully enclosed cells, so interiors only cost
  file size and assembly time.
- Strong silhouettes and 4–8 palette colors read far better than large single-color
  surfaces. Openings (arches, windows, lattices) show the cell structure.
- Features thinner than about one unit (≈1.3 cells) break up; keep spires, rails and
  cables at least 1 unit thick.
- A base (ground, island, water) anchors the build and gives the world theme a context.

## Rounds

Each round adds 20 ideas; five are selected (★) and built in parallel, then reviewed.

### Round 1

| # | Idea | Notes |
| ---: | --- | --- |
| 1 | ★ Eiffel Tower | Open iron lattice legs and arches, three platforms, park lawn |
| 2 | ★ St. Basil's Cathedral | Cluster of striped, colored onion domes on tented towers |
| 3 | ★ Taj Mahal | Marble dome, four minarets, plinth, reflecting pool and gardens |
| 4 | ★ Pirate Galleon | Curved wooden hull, three masts, sails, gun ports, stern castle |
| 5 | ★ Japanese Castle | Himeji-style stacked white keep with curved gray roofs on stone base |
| 6 | Gothic Cathedral | Twin west towers, nave, flying buttresses, rose window |
| 7 | Giza Pyramids | Three pyramids and the Sphinx on desert sand |
| 8 | Chichen Itza | Stepped Mayan pyramid with temple, stairs, jungle |
| 9 | Angkor Wat | Five lotus towers, galleries and moat |
| 10 | Lighthouse Cove | Striped lighthouse, keeper's house, rocky island, waves |
| 11 | Dutch Windmills | Row of windmills along a canal with tulip fields |
| 12 | Saturn V Launch | Rocket on its mobile launcher tower |
| 13 | Golden Gate Bridge | Red towers and cables spanning a strait between headlands |
| 14 | Sydney Opera House | White shell sails on a harbor podium |
| 15 | Parthenon | Doric temple on the Acropolis rock |
| 16 | Aircraft Carrier | Flight deck, island, parked jets |
| 17 | Steam Train | Locomotive and carriages on a viaduct |
| 18 | Neuschwanstein | Fairy-tale castle on a crag |
| 19 | Great Wall | Wall and watchtowers snaking over hills |
| 20 | Treehouse Village | Giant trees with platforms, rope bridges and huts |

Built in round 1 (all reviewed with headless `octapreview` sheets):

| Build | Cells | Review |
| --- | ---: | --- |
| `eiffel_tower` | 97,915 | Open X-braced legs, arches and three platforms; Champ de Mars, Seine and Trocadéro site |
| `st_basils_cathedral` | 118,927 | Eight patterned onion domes around the tent spire; strongest color build so far |
| `taj_mahal` | 113,335 | Onion dome, iwans, minarets, charbagh garden, mosque/jawab and gateway |
| `pirate_galleon` | 95,126 | Billowing sails, stern castle, gun decks; sits in a theme sea (`water_level` 2.4) |
| `japanese_castle` | 117,059 | Five-tier keep with curved flared roofs, three smaller keeps, moat, cherry/maple trees |

### Round 2

| # | Idea | Notes |
| ---: | --- | --- |
| 21 | Venice Grand Canal | Palazzi on both banks, Rialto bridge, gondolas |
| 22 | Mont Saint-Michel | Abbey spire crowning a walled tidal island town |
| 23 | Big Ben and Westminster | Clock tower, Gothic palace front on the Thames |
| 24 | Statue of Liberty | Full figure on the star fort pedestal |
| 25 | ★ Piazza dei Miracoli | Leaning Tower of Pisa, cathedral and round baptistery on a lawn |
| 26 | Petra Treasury | Façade carved into a rose cliff in a gorge |
| 27 | Machu Picchu | Terraces and stone houses on a ridge, Huayna Picchu peak |
| 28 | Forbidden City | Hall of Supreme Harmony on marble terraces, red walls, gold roofs |
| 29 | Hagia Sophia | Great dome, half-domes, buttresses and minarets |
| 30 | Sagrada Família | Clustered perforated spires and ornate façades |
| 31 | ★ Tower Bridge | Twin Gothic towers, raised bascules, walkways, Thames |
| 32 | Viking Harbor | Longships, longhouses and a palisade on a fjord |
| 33 | Market Town | Medieval square, church, timber houses, well |
| 34 | Space Station | Truss, modules and solar wings |
| 35 | Steampunk Airship | Envelope, gondola, propellers and fins |
| 36 | Oil Rig | Platform legs, derrick, cranes, helipad at sea |
| 37 | Easter Island | Row of moai on an ahu by the shore |
| 38 | Christ the Redeemer | Statue on the Corcovado summit |
| 39 | Arc de Triomphe | Arch with the Étoile roundabout and avenues |
| 40 | Lighthouse Harbor | Pier, boats and warehouses (merge with Lighthouse Cove) |

Also selected for round 2 from round 1's backlog: ★ Gothic Cathedral (6), ★ Neuschwanstein (18),
★ Lighthouse Cove (10).

Built in round 2:

| Build | Cells | Review |
| --- | ---: | --- |
| `gothic_cathedral` | 120,276 | West towers, rose windows, flying buttresses, flèche; Seine quays around the island |
| `neuschwanstein` | 117,010 | White Palas and slender cone-roofed towers on a forested crag with gorge, waterfall and bridge |
| `lighthouse_cove` | 91,927 | Striped lighthouse, cottage, harbor village, boats, breakwater and sea stacks in a built sea layer |
| `tower_bridge` | 105,340 | Raised bascules, walkways, suspension chains, Tower of London, HMS Belfast; long and flat |
| `piazza_dei_miracoli` | 103,854 | Leaning tower (real rotation), striped cathedral, baptistery, Camposanto and city wall |

### Round 3

| # | Idea | Notes |
| ---: | --- | --- |
| 41 | Balloon Festival | Field of hot-air balloons at different heights, baskets and tents |
| 42 | Golden Pavilion | Kinkaku-ji on its pond with pines and islands |
| 43 | Dubrovnik | Walled sea town, terracotta roofs, fortresses |
| 44 | ★ Santorini | White cube houses and blue domes cascading down a caldera cliff |
| 45 | Farmstead | Barn, silo, windmill, fields and orchard |
| 46 | Potala Palace | White and red palace terraces climbing a hill |
| 47 | Alhambra | Nasrid palaces, Court of the Lions, towers on a ridge |
| 48 | Alpine Village | Chalets and church below a Matterhorn-style peak |
| 49 | ★ Wild West Town | Main street saloons, false fronts, water tower, train station, mesa |
| 50 | Borobudur | Stepped stupa mandala with bell stupas |
| 51 | ★ Pont du Gard | Three-tier Roman aqueduct over a river gorge |
| 52 | Paddle Steamer | Mississippi riverboat with stacks and paddle wheel |
| 53 | Funfair | Ferris wheel, carousel, roller coaster loops |
| 54 | Mars Base | Domes, habitat modules, rover and lander on red terrain |
| 55 | ★ Temple of Heaven | Round triple-roofed Hall of Prayer on white marble terraces |
| 56 | Shipwreck Reef | Sunken ship among coral and rocks |
| 57 | Volcano Island | Smoking cone, lava channel, village and palms |
| 58 | Football Stadium | Bowl stands, roof ring, pitch with markings |
| 59 | Blue Mosque | Cascading domes and six minarets |
| 60 | Stilt Village | Houses on stilts over a lagoon, canoes and walkways |

Also selected from the backlog: ★ Saturn V Launch (12).

Built in round 3:

| Build | Cells | Review |
| --- | ---: | --- |
| `santorini` | 102,652 | White cube village and blue domes on a banded caldera cliff, harbor, windmills; flat cut sides |
| `wild_west_town` | 117,889 | False-front main street, church, depot and 4-4-0 train, water tower, red mesas |
| `pont_du_gard` | 107,488 | Three-tier aqueduct with real arch openings over a meandering river; diorama-cut gorge |
| `temple_of_heaven` | 101,746 | Triple blue roofs on marble terraces, courtyard gates, Danbi walkway, Imperial Vault, cypresses |
| `saturn_v_launch` | 114,041 | Saturn V with roll pattern, red umbilical tower with swing arms, flame trench, crawler, tanks |

### Round 4

| # | Idea | Notes |
| ---: | --- | --- |
| 61 | ★ Battleship | WWII dreadnought: turrets, superstructure, funnels, masts |
| 62 | Roundhouse | Locomotive turntable and engine shed with steam engines |
| 63 | Shuttle Rollout | Space Shuttle stack on the crawler near the VAB |
| 64 | ★ Floating Islands | Fantasy sky islands with waterfalls, temples, chain bridges |
| 65 | Dragon Rock | Dragon coiled on a castle crag |
| 66 | Giant Bonsai | Gnarled tree on a rock in a tray |
| 67 | Atlantis | Domed undersea city on a reef |
| 68 | Atoll Resort | Overwater bungalows around a lagoon |
| 69 | Amsterdam Canals | Gabled canal houses, bridges, boats |
| 70 | ★ Prague Old Town | Týn church spires, astronomical clock tower, square |
| 71 | Siege | Trebuchets and siege tower against a castle wall |
| 72 | Karnak Temple | Pylons, hypostyle hall, obelisks, sphinx avenue |
| 73 | Roman Forum | Temples, basilica, triumphal arch |
| 74 | Greek Theater | Epidaurus-style cavea in a hillside |
| 75 | Templo Mayor | Twin-shrine Aztec pyramid on a lake city |
| 76 | Hanseatic Harbor | Brick warehouses, cog ships, crane |
| 77 | Cyberpunk Block | Dense towers, skybridges, signs |
| 78 | ★ Meteora | Monasteries on sandstone pillars |
| 79 | Moon Base | Lander, habitats, rover on cratered ground |
| 80 | Kremlin Towers | Red brick walls and spired towers |

Also selected from the backlog: ★ Angkor Wat (9).

Built in round 4:

| Build | Cells | Review |
| --- | ---: | --- |
| `battleship` | 112,060 | Iowa-class hull with flare, triple turrets, tower bridge, funnels, catapults; sits in a theme sea |
| `floating_islands` | 96,048 | Seven sky islands with rock undersides, waterfalls, temple, pagoda, windmill, airship, bridges |
| `meteora` | 94,494 | Seven eroded pillars with monasteries, spiral stair, winch nets; valley village and river |
| `angkor_wat` | 97,939 | Five lotus towers over three galleried enclosures, naga causeway, libraries, moat |
| `prague_old_town` | 114,676 | Týn spires with gold turrets, clock tower, St Nicholas, 35 pastel gabled houses, Hus monument |

### Round 5

| # | Idea | Notes |
| ---: | --- | --- |
| 81 | Rothenburg | Half-timbered houses, towers and gate on a walled hill town |
| 82 | ★ Hobbit Village | Hill holes with round doors, gardens, mill and party tree |
| 83 | Monument Valley | Red buttes and mesas with a road and car |
| 84 | Dam and Reservoir | Curved concrete dam, spillway, power house |
| 85 | Space Elevator Base | Fantasy tether, ocean platform |
| 86 | Hollow Tree Castle | Castle carved into a giant tree |
| 87 | ★ Lunar Lander | Apollo LM on the Moon with rover, flag, craters, Earth model |
| 88 | Ice Palace | Glass and white spires on a frozen lake |
| 89 | Snow Fortress | Igloos, ice walls and towers |
| 90 | Colosseum Ruins | Ruined Colosseum with partial outer ring and hypogeum |
| 91 | Shinto Shrine | Torii gate path up a hill to a shrine |
| 92 | Tropical Beach Town | Palm promenade, pastel houses, pier |
| 93 | Coral Reef | Coral, fish, turtle and sunken amphora |
| 94 | Stonehenge | Ring of trilithons on a downland |
| 95 | Mayan City | Tikal temples rising above a jungle canopy |
| 96 | Desert Oasis | Mud-brick kasbah, palms, pool and camels |
| 97 | Cathedral Ruins | Roofless Gothic abbey with ivy |
| 98 | Medieval Harbor | Crane, cogs, warehouses, harbor towers |
| 99 | Clock Tower Town | Swiss town with Zytglogge and arcades |
| 100 | Space Shuttle Launch | Shuttle stack on the pad with service tower |

Also selected from the backlog: ★ Mont Saint-Michel (22), ★ Forbidden City (28),
★ Golden Gate Bridge (13).
