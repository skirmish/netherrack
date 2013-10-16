/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package blocks

var (
	Air                      = Block{ID: 0}
	Stone                    = Block{ID: 1, LightFiltered: 15, PlacementSound: "dig.stone"}
	Grass                    = Block{ID: 2, LightFiltered: 15, PlacementSound: "dig.grass"}
	Dirt                     = Block{ID: 3, LightFiltered: 15, PlacementSound: "dig.gravel"}
	Cobblestone              = Block{ID: 4, LightFiltered: 15, PlacementSound: "dig.stone"}
	WoodenPlanks             = Block{ID: 5, LightFiltered: 15, PlacementSound: "dig.wood"}
	Saplings                 = Block{ID: 6, PlacementSound: "dig.grass"}
	Bedrock                  = Block{ID: 7, LightFiltered: 15, PlacementSound: "dig.stone"}
	Water                    = Block{ID: 8, LightFiltered: 2}
	WaterStill               = Block{ID: 9, LightFiltered: 2}
	Lava                     = Block{ID: 10, LightEmitted: 15}
	LavaStill                = Block{ID: 11, LightEmitted: 15}
	Sand                     = Block{ID: 12, LightFiltered: 15, PlacementSound: "dig.sand"}
	Gravel                   = Block{ID: 13, LightFiltered: 15, PlacementSound: "dig.gravel"}
	GoldOre                  = Block{ID: 14, LightFiltered: 15, PlacementSound: "dig.stone"}
	IronOre                  = Block{ID: 15, LightFiltered: 15, PlacementSound: "dig.stone"}
	CoalOre                  = Block{ID: 16, LightFiltered: 15, PlacementSound: "dig.stone"}
	Log                      = Block{ID: 17, LightFiltered: 15, PlacementSound: "dig.wood"}
	Leaves                   = Block{ID: 18, LightFiltered: 1, PlacementSound: "dig.grass"}
	Sponge                   = Block{ID: 19, LightFiltered: 15, PlacementSound: "dig.grass"}
	Glass                    = Block{ID: 20, PlacementSound: "step.stone"}
	LapisLazuliOre           = Block{ID: 21, LightFiltered: 15, PlacementSound: "dig.stone"}
	LapisLazuliBlock         = Block{ID: 22, LightFiltered: 15, PlacementSound: "dig.stone"}
	Dispenser                = Block{ID: 23, LightFiltered: 15, PlacementSound: "dig.stone"}
	Sandstone                = Block{ID: 24, LightFiltered: 15, PlacementSound: "dig.stone"}
	NoteBlock                = Block{ID: 25, LightFiltered: 15, PlacementSound: "dig.stone"}
	Bed                      = Block{ID: 26}
	PoweredRail              = Block{ID: 27, PlacementSound: "step.stone"}
	DetectorRail             = Block{ID: 28, PlacementSound: "step.stone"}
	StickyPiston             = Block{ID: 29, PlacementSound: "dig.stone"}
	Cobweb                   = Block{ID: 30, PlacementSound: "dig.stone"}
	TallGrass                = Block{ID: 31, PlacementSound: "dig.grass"}
	DeadBush                 = Block{ID: 32, PlacementSound: "dig.grass"}
	Piston                   = Block{ID: 33, PlacementSound: "dig.stone"}
	PistonExtension          = Block{ID: 34}
	Wool                     = Block{ID: 35, LightFiltered: 15, PlacementSound: "dig.cloth"}
	MovedByPiston            = Block{ID: 36}
	Dandelion                = Block{ID: 37, PlacementSound: "dig.grass"}
	Poppy                    = Block{ID: 38, PlacementSound: "dig.grass"}
	BrownMushroom            = Block{ID: 39, PlacementSound: "dig.grass"}
	RedMushroom              = Block{ID: 40, PlacementSound: "dig.grass"}
	GoldBlock                = Block{ID: 41, LightFiltered: 15, PlacementSound: "dig.stone"}
	IronBlock                = Block{ID: 42, LightFiltered: 15, PlacementSound: "dig.stone"}
	DoubleSlab               = Block{ID: 43, LightFiltered: 15, PlacementSound: "dig.stone"}
	Slab                     = Block{ID: 44, LightFiltered: 1, PlacementSound: "dig.stone"}
	Bricks                   = Block{ID: 45, LightFiltered: 15, PlacementSound: "dig.stone"}
	TNT                      = Block{ID: 46, LightFiltered: 15, PlacementSound: "dig.grass"}
	Bookshelf                = Block{ID: 47, LightFiltered: 15, PlacementSound: "dig.wood"}
	MossStone                = Block{ID: 48, LightFiltered: 15, PlacementSound: "dig.stone"}
	Obsidian                 = Block{ID: 49, LightFiltered: 15, PlacementSound: "dig.stone"}
	Torch                    = Block{ID: 50, LightEmitted: 14, PlacementSound: "dig.wood"}
	Fire                     = Block{ID: 51, LightEmitted: 15}
	MonsterSpawner           = Block{ID: 52}
	OakWoodStairs            = Block{ID: 53, LightFiltered: 15, PlacementSound: "dig.wood"}
	Chest                    = Block{ID: 54, PlacementSound: "dig.wood"}
	RedstoneWire             = Block{ID: 55}
	DiamondOre               = Block{ID: 56, LightFiltered: 15, PlacementSound: "dig.stone"}
	DiamondBlock             = Block{ID: 57, LightFiltered: 15, PlacementSound: "dig.stone"}
	CraftingTable            = Block{ID: 58, LightFiltered: 15, PlacementSound: "dig.wood"}
	WheatSeeds               = Block{ID: 59, PlacementSound: "dig.grass"}
	Farmland                 = Block{ID: 60, PlacementSound: "dig.gravel"}
	Furnace                  = Block{ID: 61, LightFiltered: 15, PlacementSound: "dig.stone"}
	BurningFurnace           = Block{ID: 62, LightEmitted: 13, PlacementSound: "dig.stone"}
	SignPost                 = Block{ID: 63, PlacementSound: "dig.wood"}
	WoodenDoor               = Block{ID: 64, PlacementSound: "dig.wood"}
	Ladders                  = Block{ID: 65, PlacementSound: "dig.wood"}
	Rails                    = Block{ID: 66, PlacementSound: "step.stone"}
	CobblestoneStairs        = Block{ID: 67, LightFiltered: 15, PlacementSound: "dig.stone"}
	SignWall                 = Block{ID: 68, PlacementSound: "dig.wood"}
	Lever                    = Block{ID: 69, PlacementSound: "dig.wood"}
	StonePressurePlate       = Block{ID: 70, PlacementSound: "dig.stone"}
	IronDoor                 = Block{ID: 71}
	WoodenPressurePlate      = Block{ID: 72, PlacementSound: "dig.wood"}
	RedstoneOre              = Block{ID: 73, LightFiltered: 15, PlacementSound: "dig.stone"}
	RedstoneOreGlowing       = Block{ID: 74, LightEmitted: 9, PlacementSound: "dig.stone"}
	RedstoneTorch            = Block{ID: 75, PlacementSound: "dig.wood"}
	RedstoneTorchActive      = Block{ID: 76, LightEmitted: 7, PlacementSound: "dig.wood"}
	StoneButton              = Block{ID: 77, PlacementSound: "dig.stone"}
	Snow                     = Block{ID: 78, PlacementSound: "dig.snow"}
	Ice                      = Block{ID: 79, LightFiltered: 2, PlacementSound: "step.stone"}
	SnowBlock                = Block{ID: 80, LightFiltered: 15, PlacementSound: "dig.snow"}
	Cactus                   = Block{ID: 81, PlacementSound: "dig.cloth"}
	ClayBlock                = Block{ID: 82, LightFiltered: 15, PlacementSound: "dig.gravel"}
	SugarCane                = Block{ID: 83, PlacementSound: "dig.grass"}
	Jukebox                  = Block{ID: 84, LightFiltered: 15, PlacementSound: "dig.stone"}
	Fence                    = Block{ID: 85, PlacementSound: "dig.wood"}
	Pumpkin                  = Block{ID: 86, LightFiltered: 15, PlacementSound: "dig.wood"}
	Netherrack               = Block{ID: 87, LightFiltered: 15, PlacementSound: "dig.stone"}
	Soulsand                 = Block{ID: 88, LightFiltered: 15, PlacementSound: "dig.sand"}
	Glowstone                = Block{ID: 89, LightEmitted: 15, PlacementSound: "step.stone"}
	NetherPortal             = Block{ID: 90, LightEmitted: 11}
	JackOLantern             = Block{ID: 91, LightEmitted: 15, PlacementSound: "dig.wood"}
	CakeBlock                = Block{ID: 92, PlacementSound: "dig.cloth"}
	RedstoneRepeater         = Block{ID: 93, PlacementSound: "dig.wood"}
	RedstoneRepeaterActive   = Block{ID: 94, PlacementSound: "dig.wood"}
	StainedGlass             = Block{ID: 95, PlacementSound: "step.stone"}
	Trapdoor                 = Block{ID: 96, PlacementSound: "dig.wood"}
	MonsterEgg               = Block{ID: 97, LightFiltered: 15, PlacementSound: "dig.stone"}
	StoneBricks              = Block{ID: 98, LightFiltered: 15, PlacementSound: "dig.stone"}
	HugeBrownMushroom        = Block{ID: 99, LightFiltered: 15, PlacementSound: "dig.wood"}
	HugeRedMushroom          = Block{ID: 100, LightFiltered: 15, PlacementSound: "dig.wood"}
	IronBars                 = Block{ID: 101, PlacementSound: "step.stone"}
	GlassPane                = Block{ID: 102, PlacementSound: "step.stone"}
	Melon                    = Block{ID: 103, LightFiltered: 15, PlacementSound: "dig.wood"}
	PumpkinStem              = Block{ID: 104, PlacementSound: "dig.grass"}
	MelonStem                = Block{ID: 105, PlacementSound: "dig.grass"}
	Vines                    = Block{ID: 106, PlacementSound: "dig.grass"}
	FenceGate                = Block{ID: 107, PlacementSound: "dig.wood"}
	BrickStairs              = Block{ID: 108, LightFiltered: 15, PlacementSound: "dig.stone"}
	StoneBrickStairs         = Block{ID: 109, LightFiltered: 15, PlacementSound: "dig.stone"}
	Mycelium                 = Block{ID: 110, LightFiltered: 15, PlacementSound: "dig.grass"}
	Lilypad                  = Block{ID: 111, PlacementSound: "dig.grass"}
	NetherBrick              = Block{ID: 112, LightFiltered: 15, PlacementSound: "dig.stone"}
	NetherBrickFence         = Block{ID: 113, PlacementSound: "dig.stone"}
	NetherBrickStairs        = Block{ID: 114, LightFiltered: 15, PlacementSound: "dig.stone"}
	NetherWart               = Block{ID: 115, PlacementSound: "dig.grass"}
	EnchantmentTable         = Block{ID: 116, PlacementSound: "dig.stone"}
	BrewingStand             = Block{ID: 117, PlacementSound: "dig.stone"}
	Cauldron                 = Block{ID: 118, PlacementSound: "dig.stone"}
	EndPortal                = Block{ID: 119, LightEmitted: 15}
	EndPortalFrame           = Block{ID: 120, PlacementSound: "dig.stone"}
	EndStone                 = Block{ID: 121, LightFiltered: 15, PlacementSound: "dig.stone"}
	DragonEgg                = Block{ID: 122}
	RedstoneLamp             = Block{ID: 123, LightEmitted: 15, PlacementSound: "step.stone"}
	RedstoneLampActive       = Block{ID: 124, LightFiltered: 15, PlacementSound: "step.stone"}
	WoodenDoubleSlab         = Block{ID: 125, LightFiltered: 15, PlacementSound: "dig.wood"}
	WoodenSlab               = Block{ID: 126, LightFiltered: 1, PlacementSound: "dig.wood"}
	CocoaPod                 = Block{ID: 127}
	SandstoneStairs          = Block{ID: 128, LightFiltered: 15, PlacementSound: "dig.stone"}
	EmeraldOre               = Block{ID: 129, LightFiltered: 15, PlacementSound: "dig.stone"}
	EnderChest               = Block{ID: 130, LightEmitted: 7, PlacementSound: "dig.stone"}
	TripwireHook             = Block{ID: 131, PlacementSound: "dig.stone"}
	Tripwire                 = Block{ID: 132, PlacementSound: "dig.stone"}
	EmeraldBlock             = Block{ID: 133, LightFiltered: 15, PlacementSound: "dig.stone"}
	SpruceWoodStairs         = Block{ID: 134, LightFiltered: 15, PlacementSound: "dig.wood"}
	BirchWoodStairs          = Block{ID: 135, LightFiltered: 15, PlacementSound: "dig.wood"}
	JungleWoodStairs         = Block{ID: 136, LightFiltered: 15, PlacementSound: "dig.wood"}
	CommandBlock             = Block{ID: 137, LightFiltered: 15, PlacementSound: "dig.stone"}
	Beacon                   = Block{ID: 138, LightEmitted: 15, PlacementSound: "dig.stone"}
	CobblestoneWall          = Block{ID: 139, PlacementSound: "dig.stone"}
	FlowerPot                = Block{ID: 140, PlacementSound: "dig.stone"}
	Carrots                  = Block{ID: 141, PlacementSound: "dig.grass"}
	Potatoes                 = Block{ID: 142, PlacementSound: "dig.grass"}
	WoodenButton             = Block{ID: 143, PlacementSound: "dig.wood"}
	Head                     = Block{ID: 144, PlacementSound: "dig.stone"}
	Anvil                    = Block{ID: 145, PlacementSound: "random.anvil.land"}
	TrappedChest             = Block{ID: 146, LightFiltered: 15, PlacementSound: "dig.wood"}
	GoldPressurePlate        = Block{ID: 147, PlacementSound: "dig.stone"}
	IronPressurePlate        = Block{ID: 148, PlacementSound: "dig.stone"}
	RedstoneComparator       = Block{ID: 149, PlacementSound: "dig.wood"}
	RedstoneComparatorActive = Block{ID: 150, LightEmitted: 9, PlacementSound: "dig.wood"}
	DaylightSensor           = Block{ID: 151, PlacementSound: "dig.wood"}
	RedstoneBlock            = Block{ID: 152, LightFiltered: 15, PlacementSound: "dig.stone"}
	NetherQuartzOre          = Block{ID: 153, LightFiltered: 15, PlacementSound: "dig.stone"}
	Hopper                   = Block{ID: 154, PlacementSound: "dig.stone"}
	QuartzBlock              = Block{ID: 155, LightFiltered: 15, PlacementSound: "dig.stone"}
	QuartzStairs             = Block{ID: 156, LightFiltered: 15, PlacementSound: "dig.stone"}
	ActivatorRail            = Block{ID: 157, PlacementSound: "step.stone"}
	Dropper                  = Block{ID: 158, LightFiltered: 15, PlacementSound: "dig.stone"}
	StainedClay              = Block{ID: 159, LightFiltered: 15, PlacementSound: "dig.stone"}
	StainedGlassPane         = Block{ID: 160, PlacementSound: "step.stone"}
	HayBlock                 = Block{ID: 170, LightFiltered: 15, PlacementSound: "dig.grass"}
	Carpet                   = Block{ID: 171, PlacementSound: "dig.cloth"}
	HardenedClay             = Block{ID: 172, LightFiltered: 15, PlacementSound: "dig.stone"}
	CoalBlock                = Block{ID: 173, LightFiltered: 15, PlacementSound: "dig.stone"}
	PackedIce                = Block{ID: 174, LightFiltered: 15, PlacementSound: "step.stone"}
	LargeFlowers             = Block{ID: 175, PlacementSound: "dig.grass"}
)

type Block struct {
	ID             byte
	LightEmitted   byte
	LightFiltered  byte
	PlacementSound string
	Solid          bool
}
