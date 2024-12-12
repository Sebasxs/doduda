package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/charmbracelet/log"

	"github.com/dofusdude/ankabuffer"
	"github.com/dofusdude/doduda/ui"
	"github.com/dofusdude/doduda/unpack"
)

func unpackD2pFolder(title string, inPath string, outPath string, headless bool) {
	files := []string{}
	filepath.Walk(inPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".d2p" {
			files = append(files, path)
		}
		return nil
	})

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		os.MkdirAll(outPath, os.ModePerm)
	}

	updateProgress := make(chan bool, len(files))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ui.Progress("Unpack "+title, len(files), updateProgress, 0, true, headless)
	}()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		converted := unpack.NewD2P(f).GetFiles()
		for filename, specs := range converted {
			outFile := filepath.Join(outPath, filename)

			if filepath.Ext(filename) == ".swl" {
				log.Warnf("can not unpack swl file %s", filename)
			}

			f, err := os.Create(outFile)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			_, err = f.Write(specs["binary"].([]byte))
			if err != nil {
				log.Fatal(err)
			}
			if isChannelClosed(updateProgress) {
				os.Exit(1)
			}
		}
		updateProgress <- true
	}

	wg.Wait()
}

func DownloadImagesLauncher(hashJson *ankabuffer.Manifest, bin int, version int, dir string, headless bool) error {
	inPath := filepath.Join(dir, "tmp")
	outPath := filepath.Join(dir, "images")
	
	if version == 2 {
		fileNames := []HashFile{
			{Filename: "content/gfx/items/bitmap0.d2p", FriendlyName: "bitmaps_0.d2p"},
			{Filename: "content/gfx/items/bitmap0_1.d2p", FriendlyName: "bitmaps_1.d2p"},
			{Filename: "content/gfx/items/bitmap1.d2p", FriendlyName: "bitmaps_2.d2p"},
			{Filename: "content/gfx/items/bitmap1_1.d2p", FriendlyName: "bitmaps_3.d2p"},
			{Filename: "content/gfx/items/bitmap1_2.d2p", FriendlyName: "bitmaps_4.d2p"},
		}
		
		if err := DownloadUnpackFiles("Item Bitmaps", bin, hashJson, "main", fileNames, dir, inPath, false, "", headless, false); err != nil {
			return err
		}
		
		unpackD2pFolder("Item Bitmaps", inPath, outPath, headless)
		
		fileNames = []HashFile{
			{Filename: "content/gfx/items/vector0.d2p", FriendlyName: "vector_0.d2p"},
			{Filename: "content/gfx/items/vector0_1.d2p", FriendlyName: "vector_1.d2p"},
			{Filename: "content/gfx/items/vector1.d2p", FriendlyName: "vector_2.d2p"},
			{Filename: "content/gfx/items/vector1_1.d2p", FriendlyName: "vector_3.d2p"},
			{Filename: "content/gfx/items/vector1_2.d2p", FriendlyName: "vector_4.d2p"},
		}
		
		inPath = filepath.Join(dir, "tmp", "vector")
		outPath = filepath.Join(dir, "vector", "item")
		if err := DownloadUnpackFiles("Item Vectors", bin, hashJson, "main", fileNames, dir, inPath, false, "", headless, false); err != nil {
			return err
		}

		unpackD2pFolder("Item Vectors", inPath, outPath, headless)

		return nil
	} else if version == 3 {
		err := PullImages([]string{"stelzo/assetstudio-cli:" + ARCH}, false, headless)	
		if err != nil { return err }
		
		fileNames := []HashFile{ 
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Items/item_assets_2x.bundle", FriendlyName: "item_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Monsters/monster_assets_2x.bundle", FriendlyName: "monster_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/mount_assets_.bundle", FriendlyName: "mount_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Spells/spell_assets_2x.bundle", FriendlyName: "spell_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/alignment_assets_2x.bundle", FriendlyName: "alignment_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/challenge_assets_2x.bundle", FriendlyName: "challenges_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/companion_assets_2x.bundle", FriendlyName: "companion_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/cosmetic_assets_2x.bundle", FriendlyName: "cosmetic_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/emblem_assets_2x.bundle", FriendlyName: "emblem_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/emote_assets_2x.bundle", FriendlyName: "emote_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/job_assets_2x.bundle", FriendlyName: "job_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/preset_assets_2x.bundle", FriendlyName: "preset_images.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/smiley_assets_2x.bundle", FriendlyName: "smiley_images.imagebundle"},
		}
		err = DownloadUnpackFiles("Downloading assets", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false)
		if err != nil { return err }

		outPathUI := filepath.Join(dir, "images", "ui", "spellstates")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/Spells/spellstate_assets_all.bundle", FriendlyName: "spellstate_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading spell states", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images","ui", "achievements")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/achievement_assets_all.bundle", FriendlyName: "achievement_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading achievements", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","arena")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/arena_assets_all.bundle", FriendlyName: "arena_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading arenas", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","document")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/document_assets_all.bundle", FriendlyName: "document_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading documents", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","guidebook")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/guidebook_assets_all.bundle", FriendlyName: "guidebook_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading guidebook", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","guildrank")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/guildrank_assets_all.bundle", FriendlyName: "guildrank_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading guildranks", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","house")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/house_assets_all.bundle", FriendlyName: "house_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading houses", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","icon")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/icon_assets_all.bundle", FriendlyName: "icon_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading icons", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","illus")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/illus_assets_all.bundle", FriendlyName: "illus_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading illus", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","ornament")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/ornament_assets_all.bundle", FriendlyName: "ornament_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading ornaments", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }

		outPathUI = filepath.Join(dir, "images", "ui","suggestion")
		fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/suggestion_assets_all.bundle", FriendlyName: "suggestion_images.imagebundle"},}
		err = DownloadUnpackFiles("Downloading suggestions", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		if err != nil { return err }
		
		// outPathUI = filepath.Join(dir, "images", "worldmaps")
		// fileNames = []HashFile{ {Filename: "Dofus_Data/StreamingAssets/Content/Picto/Worldmaps/worldmap_assets__4ba03324f4420d542d1ee6d3f566f3d1.bundle", FriendlyName: "worldmap_images.imagebundle"},}
		// err = DownloadUnpackFiles("Downloading worldmaps", bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
		// if err != nil { return err }
		
		feedbacks := make(chan string)
		
		var feedbackWg sync.WaitGroup
		feedbackWg.Add(1)
		go func() {
			defer feedbackWg.Done()
			ui.Spinner("Images", feedbacks, false, headless)
			}()
			
			defer func() {
			close(feedbacks)
			feedbackWg.Wait()
			}()
			
			feedbacks <- "cleaning"
			
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "items", "2x"), filepath.Join(dir, "images", "items"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "monsters", "2x"), filepath.Join(dir, "images", "monsters"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "mounts", "big"), filepath.Join(dir, "images", "ui", "mounts"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "spells", "2x"), filepath.Join(dir, "images", "ui", "spells"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "alignments", "2x"), filepath.Join(dir, "images", "ui","alignments"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "challenges", "2x"), filepath.Join(dir, "images", "ui","challenges"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "companions", "2x"), filepath.Join(dir, "images", "ui","companions"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "cosmetics", "2x"), filepath.Join(dir, "images", "ui","cosmetics"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "emblems", "big"), filepath.Join(dir, "images", "ui","emblems"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "emotes", "2x"), filepath.Join(dir, "images", "ui","emotes"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "jobs", "2x"), filepath.Join(dir, "images", "ui","jobs"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "presets", "2x"), filepath.Join(dir, "images", "ui","presets"))
		if err != nil { return err }
		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "smilies", "2x"), filepath.Join(dir, "images", "ui","smilies"))
		if err != nil { return err }

		err = os.RemoveAll(filepath.Join(outPath, "Assets"))
		if err != nil { return err }
		err = os.RemoveAll(filepath.Join(outPath, "monster_images.imagebundle"))
		if err != nil { return err }
		err = os.RemoveAll(filepath.Join(outPath, "mount_images.imagebundle"))
		if err != nil { return err }
		err = os.RemoveAll(filepath.Join(outPath, "spell_images.imagebundle"))
		if err != nil { return err }
		err = os.RemoveAll(filepath.Join(outPath, "spellstate_images.imagebundle"))
		if err != nil { return err }

		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
