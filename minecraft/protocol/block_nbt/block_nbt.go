package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// 描述 方块实体 的通用接口
type BlockNBT interface {
	ID() string               // 返回该 方块实体 的 ID
	Marshal(io protocol.IO)   // 解码或编码为二进制的平铺型 __tag NBT
	ToNBT() map[string]any    // 将该 方块实体 所记的数据转换为 NBT
	FromNBT(x map[string]any) // 将 x 指代的 NBT 加载到该 方块实体
}

// 以下列出了各个 方块实体 的 ID
const (
	IDChiseledBookshelf = "ChiseledBookshelf"
	IDDayLightDetector  = "DaylightDetector"
	IDEndPortal         = "EndPortal"
	IDSculkCatalyst     = "SculkCatalyst"
	IDSporeBlossom      = "SporeBlossom"

	IDBrushableBlock = "BrushableBlock"
	IDDecoratedPot   = "DecoratedPot"

	IDBanner          = "Banner"
	IDBeacon          = "Beacon"
	IDBed             = "Bed"
	IDBeehive         = "Beehive"
	IDBell            = "Bell"
	IDBrewingStand    = "BrewingStand"
	IDCampfire        = "Campfire"
	IDCommandBlock    = "CommandBlock"
	IDComparator      = "Comparator"
	IDConduit         = "Conduit"
	IDCauldron        = "Cauldron"
	IDEnchantingTable = "EnchantTable"
	IDFlowerPot       = "FlowerPot"
	IDHopper          = "Hopper"
	IDJigsaw          = "JigsawBlock"
	IDJukebox         = "Jukebox"
	IDLectern         = "Lectern"
	IDLodestone       = "Lodestone"
	IDMobSpawner      = "MobSpawner"
	IDMovingBlock     = "MovingBlock"
	IDNetherReactor   = "NetherReactor"
	IDNoteBlock       = "Music"
	IDPiston          = "PistonArm"
	IDSkull           = "Skull"
	IDStructureBlock  = "StructureBlock"

	IDSign        = "Sign"
	IDHangingSign = "HangingSign"

	IDSculkSensor           = "SculkSensor"
	IDCalibratedSculkSensor = "CalibratedSculkSensor"
	IDSculkShrieker         = "SculkShrieker"

	IDFurnace      = "Furnace"
	IDBlastFurnace = "BlastFurnace"
	IDSmoker       = "Smoker"

	IDChest      = "Chest"
	IDBarrel     = "Barrel"
	IDEnderChest = "EnderChest"
	IDShulkerBox = "ShulkerBox"

	IDDispenser = "Dispenser"
	IDDropper   = "Dropper"

	IDFrame     = "ItemFrame"
	IDGlowFrame = "GlowItemFrame"

	IDChemistryTable = "ChemistryTable"
	IDModBlock       = "ModBlock"
)

// 返回一个方块实体池，
// 其中包含了 方块实体 的 ID 到其对应 方块实体 的映射
func NewPool() map[string]BlockNBT {
	return map[string]BlockNBT{
		IDBanner:                &Banner{},
		IDBarrel:                &Barrel{},
		IDBeacon:                &Beacon{},
		IDBed:                   &Bed{},
		IDBeehive:               &Beehive{},
		IDBell:                  &Bell{},
		IDBlastFurnace:          &BlastFurnace{},
		IDBrewingStand:          &BrewingStand{},
		IDBrushableBlock:        &BrushableBlock{},
		IDCalibratedSculkSensor: &CalibratedSculkSensor{},
		IDCampfire:              &Campfire{},
		IDCauldron:              &Cauldron{},
		IDChemistryTable:        &ChemistryTable{},
		IDChest:                 &Chest{},
		IDChiseledBookshelf:     &ChiseledBookshelf{},
		IDCommandBlock:          &CommandBlock{},
		IDComparator:            &Comparator{},
		IDConduit:               &Conduit{},
		IDDayLightDetector:      &DayLightDetector{},
		IDDecoratedPot:          &DecoratedPot{},
		IDDispenser:             &Dispenser{},
		IDDropper:               &Dropper{},
		IDEnchantingTable:       &EnchantingTable{},
		IDEndPortal:             &EndPortal{},
		IDEnderChest:            &EnderChest{},
		IDFlowerPot:             &FlowerPot{},
		IDFurnace:               &Furnace{},
		IDGlowFrame:             &GlowFrame{},
		IDHangingSign:           &HangingSign{},
		IDHopper:                &Hopper{},
		IDFrame:                 &Frame{},
		IDJigsaw:                &Jigsaw{},
		IDJukebox:               &Jukebox{},
		IDLectern:               &Lectern{},
		IDLodestone:             &Lodestone{},
		IDMobSpawner:            &MobSpawner{},
		IDModBlock:              &ModBlock{},
		IDMovingBlock:           &MovingBlock{},
		IDNetherReactor:         &NetherReactor{},
		IDNoteBlock:             &NoteBlock{},
		IDPiston:                &Piston{},
		IDSculkCatalyst:         &SculkCatalyst{},
		IDSculkSensor:           &SculkSensor{},
		IDSculkShrieker:         &SculkShrieker{},
		IDShulkerBox:            &ShulkerBox{},
		IDSign:                  &Sign{},
		IDSkull:                 &Skull{},
		IDSmoker:                &Smoker{},
		IDSporeBlossom:          &SporeBlossom{},
		IDStructureBlock:        &StructureBlock{},
	}
}