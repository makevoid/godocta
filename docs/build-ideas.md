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
| 82 | ★ Hillside Burrows | Burrow homes with round doors, gardens, mill and party tree |
| 83 | Monument Valley | Red buttes and mesas with a road and car |
| 84 | Dam and Reservoir | Curved concrete dam, spillway, power house |
| 85 | Space Elevator Base | Fantasy tether, ocean platform |
| 86 | Hollow Tree Castle | Castle carved into a giant tree |
| 87 | ★ Lunar Landing | Apollo LM on the Moon with rover, flag, craters, Earth model |
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

Built in round 5 (Hobbit Village was renamed Hillside Burrows to avoid a trademarked name;
Mont Saint-Michel's title drops the hyphen so its world key matches):

| Build | Cells | Review |
| --- | ---: | --- |
| `golden_gate_bridge` | 118,438 | Orange towers, catenary cables and suspenders, truss deck with traffic, headlands, container ship |
| `lunar_landing` | 99,214 | Gold-foil LM, rover, astronauts, flag, experiments, cratered regolith; free-floating Earth |
| `hillside_burrows` | 85,147 | Round-door burrow homes in turf hills, great oak, mill pond, bridge, party marquee |
| `mont_saint_michel` | 82,154 | Abbey and spire on a granite island, ramparts, Grande Rue houses, tidal flats and causeway |
| `forbidden_city` | 114,708 | Meridian Gate, Golden Water bridges, three halls on marble terraces, gold roofs, red walls, moat |

### Round 6

| # | Idea | Notes |
| ---: | --- | --- |
| 101 | Bamboo Pagoda Garden | Pagoda, koi pond, moon bridge, bamboo grove |
| 102 | Harbor Lighthouse Fort | Star fort with bastions on a headland |
| 103 | Swiss Railway Viaduct | Landwasser-style curved viaduct into a tunnel |
| 104 | Medieval Monastery | Cloister, church, scriptorium, gardens |
| 105 | Sky City Dock | Airships moored at a floating city's docks |
| 106 | Gas Giant Station | Sci-fi ring station (free-floating) |
| 107 | Grain Mill Farm | Water mill, granary and fields |
| 108 | Winter Village | Snow-roofed chalets, frozen pond, pines |
| 109 | Pagoda Mountain | Mountain temple with stairs through mist |
| 110 | Castle Ruins | Ruined keep, broken walls, ivy, crows |
| 111 | Volcano Temple | Temple on a volcano with lava flows |
| 112 | Dutch Harbor Town | Gabled houses, drawbridge, tjalk barges |
| 113 | Train Station Hall | Iron-and-glass shed, platforms, locomotives |
| 114 | Space Port | Launch towers, hangars, starships |
| 115 | Underwater Base | Domes and tubes on the seabed |
| 116 | Rice Terraces | Banaue-style terraces with huts |
| 117 | Mississippi Riverfront | Paddle steamers at a levee town |
| 118 | Royal Observatory | Domes and telescopes on a hill |
| 119 | Chinese Garden | Courtyards, rockery, pavilions |
| 120 | Cliff Palace | Mesa Verde dwellings in a canyon alcove |

Selected from the backlog: ★ Sydney Opera House (14), ★ Venice Grand Canal (21),
★ Machu Picchu (27), ★ Hagia Sophia (29), ★ Funfair (53).

Built in round 6:

| Build | Cells | Review |
| --- | ---: | --- |
| `sydney_opera_house` | 97,090 | Interlocking white shell sails with chevron ribs, glass mouths, stepped podium, harbor and ferry |
| `funfair` | 91,161 | Ferris wheel, looping coaster, carousel, helter-skelter, striped tents, swan pond, entrance arch |
| `hagia_sophia` | 107,963 | Dome cascade, window drum, buttresses, four minarets, türbes; dark interiors behind openings |
| `machu_picchu` | 90,608 | Terraced citadel on its saddle, Huayna Picchu, Urubamba loop, thatched houses, llamas |
| `venice_grand_canal` | 91,760 | Curved canal of Gothic palazzi, Rialto with shops, campanili, gondolas, side canals |

### Round 7

| # | Idea | Notes |
| ---: | --- | --- |
| 121 | Lost Jungle Temple | Overgrown stepped temple with roots, waterfall |
| 122 | Harbour Crane Port | Container cranes, stacks, ship at berth |
| 123 | Gingerbread Village | Candy houses, icing roofs, lollipop trees |
| 124 | ★ Wizard Academy | Fantasy castle school with towers on a lake cliff |
| 125 | Desert Fortress | Mud-brick ksar with towers and gate |
| 126 | Chinese Junk Fleet | Junks with battened sails in a harbor |
| 127 | Vineyard Chateau | Loire château with vineyards and river |
| 128 | Medieval Bridge Town | Houses built on a bridge (Ponte Vecchio) |
| 129 | Alpine Ski Resort | Lodge, lifts, slopes and pines |
| 130 | Pagoda of Pagodas | Shwedagon-style golden stupa with temples |
| 131 | Lighthouse Rock | Bell Rock style tower on a reef at sea |
| 132 | Tree of Life | Giant fantasy tree with a city in its branches |
| 133 | Steamboat Wharf | River landing with warehouses and steamers |
| 134 | Ancient Library | Library of Alexandria style colonnaded hall |
| 135 | Ice Hotel | Ice blocks, arches and aurora-lit snowfield |
| 136 | Space Telescope | Hubble-style telescope with solar wings (floating) |
| 137 | Aztec Ballcourt | I-shaped court with rings and temples |
| 138 | Abbey on the Rock | Skellig-style beehive huts on a sea crag |
| 139 | Desert Railway | Steam train crossing a trestle in a canyon |
| 140 | Toy Train Set | Track loop, station, tunnel mountain, village |

Selected from the backlog: ★ Statue of Liberty (24), ★ Petra Treasury (26),
★ Sagrada Família (30), ★ Viking Harbor (32).

Built in round 7:

| Build | Cells | Review |
| --- | ---: | --- |
| `sagrada_familia` | 114,691 | Eighteen perforated spires with mosaic finials, Jesus tower cross, dripping Nativity portals, pond plaza |
| `statue_of_liberty` | 108,124 | Robed figure with torch, crown and tablet on the granite pedestal and star fort; patina reads pale |
| `petra_treasury` | 85,026 | Two-storey façade with tholos carved in banded rose cliffs, plaza, Siq, camels and tombs |
| `wizard_academy` | 115,628 | Gothic school of cone-roofed towers on a loch cliff, clock tower, greenhouse, viaduct train |
| `viking_harbor` | 115,269 | Fjord walls with waterfalls, four longships, palisaded turf longhouses, great hall, stave church |

### Round 8

| # | Idea | Notes |
| ---: | --- | --- |
| 141 | Medieval Fair | Jousting lists, pavilions, stands and banners |
| 142 | Coastal Fortress | Vauban citadel by a harbor town |
| 143 | Pagoda Bridge Village | Water town (Zhouzhuang) with arched bridges |
| 144 | Wind Farm Coast | Wind turbines on cliffs and offshore |
| 145 | Rocket Garden | Display of historic rockets around a museum |
| 146 | Haunted Manor | Victorian mansion, graveyard, dead trees |
| 147 | Arctic Research Station | Huts, radar dome, icebreaker in floes |
| 148 | Carnival Cruise Port | Terminal, cruise ship, palms |
| 149 | Bavarian Village | Onion-dome church, painted houses, maypole |
| 150 | Cathedral Square | Russian kremlin cathedral cluster with gold domes |
| 151 | Samurai Village | Thatched farmhouses, rice paddies, shrine |
| 152 | Floating Market | Boats laden with fruit along stilt houses |
| 153 | Cliffside Monastery | Tiger's Nest style temple on a sheer cliff |
| 154 | Lighthouse Pier | Victorian seaside pier with pavilions |
| 155 | Trojan Horse | Giant wooden horse before city walls |
| 156 | Mammoth Camp | Ice-age hunters' tents, mammoths, tundra |
| 157 | Dinosaur Valley | Sauropods and a T. rex among ferns and a volcano |
| 158 | Ruined Aqueduct City | Overgrown Roman town |
| 159 | Pagoda Waterfall | Temple beside a tall waterfall in mountains |
| 160 | Rooftop City | Dense Mediterranean rooftops with domes |

Selected from the backlog: ★ Giza Pyramids (7), ★ Chichen Itza (8), ★ Big Ben (23),
★ Aircraft Carrier (16), ★ Treehouse Village (20).

Built in round 8:

| Build | Cells | Review |
| --- | ---: | --- |
| `big_ben` | 104,561 | Clock tower with dials and gold spire, Palace river front, Victoria Tower, green bridge with buses |
| `aircraft_carrier` | 101,583 | Angled flight deck with markings, parked jets, elevators, island with radars; mostly dark grey |
| `treehouse_village` | 113,319 | Five giant trees with ringed platforms, huts, rope bridges, mushrooms, stream and mill |
| `giza_pyramids` | 114,019 | Three course-stepped pyramids, Khafre's cap, Sphinx enclosure, temples, mastabas, caravan; low contrast |
| `chichen_itza` | 105,847 | El Castillo with serpent stairs, Warriors temple, ball court, Caracol, sacbe and cenote, jungle ring |

### Round 9

| # | Idea | Notes |
| ---: | --- | --- |
| 161 | Tulip Fields | Windmill, canal and striped flower fields |
| 162 | Castle Siege Camp | Besieging army tents and engines |
| 163 | Monastery Island | Skellig Michael beehive huts on a sea crag |
| 164 | Seaside Boardwalk | Coney Island style boardwalk and rides |
| 165 | Mining Town | Headframe, rail carts, wooden town in hills |
| 166 | Pyramid of the Sun | Teotihuacan avenue and pyramids |
| 167 | Fishing Village | Huts, nets, boats and a pier |
| 168 | Thatched Cottage Village | Cotswold cottages, church, stream |
| 169 | Neon Arcade Street | Asian night-market alley with signs |
| 170 | Canal Lock Flight | Narrowboats on a lock staircase |
| 171 | Imperial Palace Garden | Japanese garden with castle keep |
| 172 | Clocktower Square | Italian piazza with a clock tower and arcades |
| 173 | Eagle's Nest | Mountain-top chalet over cliffs |
| 174 | Polar Lighthouse | Lighthouse and huts on an ice shelf |
| 175 | Rainforest Canopy Walk | Suspension bridges between emergent trees |
| 176 | Space Shuttle Hangar | Shuttle in an open hangar with gantries |
| 177 | Pirate Cove | Hidden cave harbor with a ship and shacks |
| 178 | Old West Fort | Log stockade with blockhouses and a river |
| 179 | Canyon Bridge | Steel arch bridge over a deep gorge |
| 180 | Cloud Temple | Temple on pillars above a sea of clouds |

Selected from the backlog: ★ Parthenon (15), ★ Great Wall (19), ★ Dubrovnik (43),
★ Potala Palace (46), ★ Steampunk Airship (35).

Built in round 9:

| Build | Cells | Review |
| --- | ---: | --- |
| `potala_palace` | 115,374 | Red Palace with gold roofs over stepped white wings, bastions, Shöl village, chorten |
| `steampunk_airship` | 87,246 | Striped envelope, fins, gondola with paddle wheels, engines, lattice mast on a hill; ship hangs from the mast |
| `parthenon` | 100,193 | Ruined peristyle on the Acropolis rock, Propylaea, Erechtheion, Odeon; Erechtheion hidden in hero view |
| `great_wall` | 101,107 | S-curving wall over autumn ridges, 8 varied towers, pass fortress, village |
| `dubrovnik` | 100,370 | Walled terracotta town on the sea, Minčeta, Lovrijenac, harbor with carrack; modest in hero frame |

### Round 10

| # | Idea | Notes |
| ---: | --- | --- |
| 181 | Lighthouse Island Village | Fishing hamlet around a lighthouse on a rock |
| 182 | Hanging Gardens | Terraced ziggurat gardens of Babylon |
| 183 | Ishtar Gate | Blue glazed gate and processional way |
| 184 | Hot Spring Onsen | Japanese inn with steaming pools in snow |
| 185 | Medieval Windmill Hill | Post mills on a hill above a village |
| 186 | Submarine Dock | Submarine in a dry dock with cranes |
| 187 | Zeppelin Hangar | Huge hangar with an airship half inside |
| 188 | Pagoda Temple Complex | Horyu-ji style pagoda, hall and cloister |
| 189 | Brooklyn Bridge | Gothic stone towers and cable web over a river |
| 190 | Sphinx Temple | Sphinx close-up with temple ruins |
| 191 | Coastal Castle | Welsh castle (Harlech) above a beach |
| 192 | Kasbah Village | Ait Benhaddou mud-brick hill town |
| 193 | Grand Canyon Lodge | Lodge on a canyon rim with layered strata |
| 194 | Lunar Rover Camp | Rovers and a habitat in a crater |
| 195 | Elven Forest City | Graceful towers among giant trees |
| 196 | Clipper Ship | Tall sailing clipper under full sail at sea |
| 197 | Trireme Fleet | Greek warships rowing by a harbor |
| 198 | Castle Gatehouse | Massive gatehouse with drawbridge and moat |
| 199 | Tea Plantation | Terraced tea rows with a colonial bungalow |
| 200 | Ziggurat of Ur | Stepped mud-brick ziggurat with stairs |

Selected from the backlog: ★ Steam Train (17), ★ Christ the Redeemer (38), ★ Alhambra (47),
★ Borobudur (50), ★ Mars Base (54).

Built in round 10:

| Build | Cells | Review |
| --- | ---: | --- |
| `borobudur` | 88,365 | Redented square galleries, 72 bell stupas, central stupa on a palm park; mostly one grey |
| `mars_base` | 85,704 | Glass domes with crops, hub and habitats, steel rocket, comms mast, rovers, solar farm |
| `steam_train` | 82,006 | Blue express on a curved 14-arch viaduct, plume, loch, kirk and heather hills |
| `christ_the_redeemer` | 92,741 | Arms-outstretched statue on its pedestal atop a granite peak, cog railway, cove and Sugarloaf |
| `alhambra` | 86,775 | Red fortress walls and towers, Comares, Charles V palace, courts, Generalife, Darro river |

### Round 11

| # | Idea | Notes |
| ---: | --- | --- |
| 201 | Harbor Drawbridge | Bascule bridge raised for a tall ship |
| 202 | Royal Crescent | Georgian crescent terrace with a lawn |
| 203 | Spanish Galleon Battle | Two ships trading broadsides |
| 204 | Space Colony Cylinder | Cutaway O'Neill cylinder (floating) |
| 205 | Lake Pavilion | Chinese pavilion with zigzag bridge on a lake |
| 206 | Teahouse Mountain | Tea house on stilts on a cliff |
| 207 | Abandoned Factory | Brick mill, chimney, overgrown yard |
| 208 | Observatory Dome | Single large telescope dome on a peak |
| 209 | Viking Longhouse Farm | Turf-roofed longhouse and fields |
| 210 | Water Temple | Balinese temple on a lake (Ulun Danu) |
| 211 | Pirate Fortress | Caribbean fort with cannons over a bay |
| 212 | Medieval Joust | Tilt yard and stands (merge with Medieval Fair) |
| 213 | Elephant Caravan | Decorated elephants with howdahs on a road |
| 214 | Totem Village | Northwest Coast plank houses and totems |
| 215 | Mesa Pueblo | Taos-style adobe terraces |
| 216 | Clock Workshop | Giant cutaway clock mechanism |
| 217 | Ferris Wheel Pier | Large wheel at the end of a pier |
| 218 | Dragon Boat Race | Boats with dragon heads on a river |
| 219 | Stave Church | Norwegian stave church on a hill |
| 220 | Crystal Cave | Cave with giant crystals and a pool |

Selected: ★ Easter Island (37), ★ Golden Pavilion (42), ★ Volcano Island (57),
★ Colosseum Ruins (90), ★ Haunted Manor (146).

Built in round 11:

| Build | Cells | Review |
| --- | ---: | --- |
| `volcano_island` | 90,426 | Smoking cone with lava channel to the sea, jungle, hut village, striped lighthouse, reef |
| `easter_island` | 97,951 | 15 moai with pukao on Ahu Tongariki, Rano Raraku crater lake, lava coast; wide site |
| `colosseum_ruins` | 117,816 | Full-height north façade, ruined cavea, hypogeum, Stern buttress, Arch of Constantine, pines |
| `haunted_manor` | 103,613 | Navy Second Empire mansion with tower and turret, graveyard, dead trees, pond; floating bats/ghosts |
| `golden_pavilion` | 84,606 | Gold pavilion on its mirror pond, pine islands, autumn maples, tea house, wooded hill |

### Round 12

| # | Idea | Notes |
| ---: | --- | --- |
| 221 | Rialto Market | Venetian fish market and bridge (overlaps Venice) |
| 222 | Lighthouse Storm | Wave-battered lighthouse (static spray) |
| 223 | Bell Tower Plaza | Venetian campanile and loggia on a square |
| 224 | Gondola Lift Peak | Cable car station on an alpine peak |
| 225 | Dutch Polder | Windmills pumping along dykes |
| 226 | Kowloon Walled City | Dense stacked blocks (overlaps Cyberpunk Block) |
| 227 | Medieval Castle Keep | Square Norman keep with bailey |
| 228 | Observatory Island | Radio telescope dish on a rocky isle |
| 229 | Pagoda Pine Island | Small island with a pagoda (overlaps) |
| 230 | Desert Caravanserai | Courtyard inn with camels |
| 231 | Brutalist Campus | Concrete towers and plazas |
| 232 | Salt Flats Rally | Desert racing camp with trucks and tents |
| 233 | Arctic Ice Ship | Icebreaker in pack ice |
| 234 | Tuscan Hill Town | San Gimignano towers on a hill |
| 235 | Japanese Shrine Gate | Itsukushima floating torii and shrine |
| 236 | Inca Rope Bridge | Grass bridge over a gorge |
| 237 | Spanish Mission | Adobe mission church and courtyard |
| 238 | Grain Elevator Town | Prairie elevators and railway |
| 239 | Floating Pagoda | Fantasy pagoda on a floating rock |
| 240 | Old Mill Pond | Water mill with wheel and pond |

Selected: ★ Kremlin Towers (80), ★ Amsterdam Canals (69), ★ Cliffside Monastery (153),
★ Clipper Ship (196), ★ Karnak Temple (72).

Built in round 12:

| Build | Cells | Review |
| --- | ---: | --- |
| `clipper_ship` | 102,269 | Black-hulled tea clipper under full square rig, heeling, bow wave and wake |
| `tigers_nest` | 106,066 | White and red temples with gold roofs in a cliff alcove, waterfall, prayer flags; blocky massif from the side |
| `kremlin` | 92,407 | Red walls and green-tented towers, Spasskaya, gold-domed cathedrals, Red Square, river |
| `amsterdam_canals` | 95,145 | Gabled canal houses, two curved canals, bridges and boats, Westerkerk crown tower |
| `karnak_temple` | 99,805 | Sphinx avenue, First Pylon, hypostyle hall, obelisks, sacred lake; low contrast, hall reads as a block |

### Round 13

| # | Idea | Notes |
| ---: | --- | --- |
| 241 | Harbour Fish Market | Covered market hall by a quay |
| 242 | Desert Monastery | Saint Catherine's walls under a granite peak |
| 243 | Skyscraper Under Construction | Steel frame, cranes, scaffolds |
| 244 | Bamboo Forest Temple | Shrine in tall bamboo |
| 245 | Royal Yacht Harbor | Superyachts at a marina below a casino |
| 246 | Celtic Hillfort | Ringed earthworks, roundhouses, palisade |
| 247 | Lighthouse Causeway | Lighthouse reached by a tidal causeway |
| 248 | Viking Ship Burial | Ship in a mound being covered |
| 249 | Tropical Treehouse Resort | Huts in rainforest canopy (overlaps) |
| 250 | Matterhorn Hut | Alpine hut below a pyramidal peak |
| 251 | Baroque Palace Garden | Versailles-like parterres and fountains |
| 252 | Mughal Fort | Red Fort walls and pavilions |
| 253 | Pueblo Cliff Town | (see Cliff Palace) |
| 254 | Steel Mill | Blast furnace, stacks, rail yard |
| 255 | Highland Castle | Eilean Donan on a loch islet with bridge |
| 256 | Bridge of Sighs | Oxford/Venice covered bridge scene |
| 257 | Rainbow Village | Colourful hillside houses (Cinque Terre) |
| 258 | Tibetan Stupa | Boudhanath stupa with eyes |
| 259 | Shipyard Slipway | Ship on stocks with cranes |
| 260 | Glacier Lagoon | Icebergs and a blue glacier front |

Selected: ★ Rothenburg (81), ★ Monument Valley (83), ★ Cliff Palace (120),
★ Ishtar Gate (183), ★ Stave Church (219).

Built in round 13:

| Build | Cells | Review |
| --- | ---: | --- |
| `rothenburg` | 96,038 | Walled half-timbered town on a plateau, Rathaus tower, St. Jakob spires, vineyards, double-decker bridge |
| `monument_valley` | 89,188 | Mittens and Merrick Butte, straight road with a car, hogan and corral; mostly red-orange |
| `cliff_palace` | 113,865 | Masonry rooms, Square Tower and kivas in a sandstone alcove; boxy mesa, village fills only the lower alcove |
| `ishtar_gate` | 100,895 | Navy double gate with gold aurochs and dragons, Processional Way, mud-brick walls, ziggurat, Euphrates |
| `stave_church` | 94,150 | Tiered tarred roofs with gold dragon heads, bell tower, turf-roofed stabbur, waterfall, fjord |

### Round 14

| # | Idea | Notes |
| ---: | --- | --- |
| 261 | Kinderdijk Windmills | Row of polder mills along dyked canals |
| 262 | Offshore Oil Rig | Platform legs, derrick, flare boom, supply boat |
| 263 | Blue Mosque Domes | Cascading domes and six minarets |
| 264 | Eilean Donan | Highland castle on a loch islet with an arched bridge |
| 265 | Cinque Terre Cove | Pastel houses stacked above a tiny harbour |
| 266 | Sky Tram Station | Cable car over a gorge between two stations |
| 267 | Ice Hotel | Glowing ice blocks and arched halls |
| 268 | Bavarian Beer Garden | Maypole, chestnut trees, timber hall |
| 269 | Roman Aqueduct Town | Segovia arches over a hill town |
| 270 | Space Elevator Base | Tether anchor on an ocean platform |
| 271 | Medieval Siege Camp | Trebuchets and tents outside walls |
| 272 | Cathedral Ruins | Roofless abbey with grassy nave (Tintern) |
| 273 | Hot Spring Onsen | Ryokan with steaming pools in snow |
| 274 | Sea Fort | Round island fort (Spitbank) |
| 275 | Lantern Festival Street | Lanterns over a Chinese old street |
| 276 | Racing Circuit Pit Lane | Grandstand, pit garages, cars |
| 277 | Giant's Causeway | Basalt column steps into the sea |
| 278 | Monastery on a Lake | Island monastery (Bled) |
| 279 | Floating Market | Boats along a Thai canal |
| 280 | Arctic Research Station | Huts, radar dome, snowcats |

Selected: ★ Eilean Donan (264), ★ Blue Mosque Domes (263), ★ Offshore Oil Rig (262),
★ Cinque Terre Cove (265), ★ Kinderdijk Windmills (261).
