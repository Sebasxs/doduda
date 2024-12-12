package main

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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


func moveFilesToParentFolder(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(dir)

	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		newFilePath := filepath.Join(parentDir, file.Name())
		err := os.Rename(filePath, newFilePath)
		if err != nil {
			return err
		}
	}

	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}

func checkImageDimensions(imagePath string, dim int) (bool, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Error("Error opening file", "err", err)
		return false, err
	}
	defer file.Close()

	img, err := png.DecodeConfig(file)
	if err != nil {
		log.Errorf("Error decoding image, skipping %s\n", imagePath)
		return false, nil
	}

	return img.Width != dim && img.Height != dim, nil
}

func isDuplicatedName(filename string) bool {
	return strings.Contains(filename, "_#")
}

func cleanImages(dir string, dim int, patternExcluded *regexp.Regexp) error {
	deletedCount := 0
	renamedCount := 0

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	total := len(files)
	assetType := strings.Split(dir, "images")[1]

	for _, file := range files {
		imagePath := filepath.Join(dir, file.Name())

		if file.IsDir() || filepath.Ext(imagePath) != ".png" {
			return nil
		}

		fitDimensions, err := checkImageDimensions(imagePath, dim)
		if err != nil {
			log.Printf("Error checking dimensions of %s: %v", file.Name(), err)
			return err
		}

		if dim > 0 && !fitDimensions {
			err := os.Remove(imagePath)
			if err != nil {
				log.Printf("Error deleting %s: %v", file.Name(), err)
			} else {
				deletedCount++
			}
		} else if isDuplicatedName(file.Name()) {
			newName := regexp.MustCompile(`_#\d+`).ReplaceAllString(file.Name(), "")
			cleanImagePath := filepath.Join(dir, newName)

			_, err := os.Stat(cleanImagePath)
			fileExists := !os.IsNotExist(err)

			if fileExists && patternExcluded != nil && patternExcluded.MatchString(file.Name()) {
				return nil
			}

			err = os.Rename(imagePath, cleanImagePath)
			if err != nil {
				log.Printf("Error renaming %s: %v", file.Name(), err)
			} else {
				renamedCount++
			}
		}
	}
	
	fmt.Printf("Renamed: %d, Deleted: %d from %d in %s\n", renamedCount, deletedCount, total, assetType)
	return err
}

func DownloadImagesLauncher(hashJson *ankabuffer.Manifest, bin int, version int, dir string, headless bool) error {
	inPath := filepath.Join(dir, "tmp")
	outPath := filepath.Join(dir, "images")
	monstersPath := filepath.Join(dir, "images", "monsters")
	uiPath := filepath.Join(dir, "images", "ui")
	itemsPath := filepath.Join(dir, "images", "items")
	
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

		uiPaths := map[string]string{
			"arena":        "Dofus_Data/StreamingAssets/Content/Picto/UI/arena_assets_all.bundle",
			"achievements": "Dofus_Data/StreamingAssets/Content/Picto/UI/achievement_assets_all.bundle",
			"document":     "Dofus_Data/StreamingAssets/Content/Picto/UI/document_assets_all.bundle",
			"guidebook":    "Dofus_Data/StreamingAssets/Content/Picto/UI/guidebook_assets_all.bundle",
			"guildrank":    "Dofus_Data/StreamingAssets/Content/Picto/UI/guildrank_assets_all.bundle",
			"house":        "Dofus_Data/StreamingAssets/Content/Picto/UI/house_assets_all.bundle",
			"icon":         "Dofus_Data/StreamingAssets/Content/Picto/UI/icon_assets_all.bundle",
			"illus":        "Dofus_Data/StreamingAssets/Content/Picto/UI/illus_assets_all.bundle",
			"ornament":     "Dofus_Data/StreamingAssets/Content/Picto/UI/ornament_assets_all.bundle",
			"spellstates":  "Dofus_Data/StreamingAssets/Content/Picto/Spells/spellstate_assets_all.bundle",
			"suggestion":   "Dofus_Data/StreamingAssets/Content/Picto/UI/suggestion_assets_all.bundle",
			// "worldmaps":    "Dofus_Data/StreamingAssets/Content/Picto/Worldmaps/worldmap_assets__4ba03324f4420d542d1ee6d3f566f3d1.bundle",
		}

		for key, path := range uiPaths {
			outPathUI := filepath.Join(uiPath, key)
			fileNames := []HashFile{{Filename: path, FriendlyName: key + "_images.imagebundle"}}
			err = DownloadUnpackFiles("Downloading "+key, bin, hashJson, "picto", fileNames, dir, outPathUI, true, "", headless, false)
			if err != nil {
				return err
			}
		}
		
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
		
		renamePaths := map[string]string{
			"items":                				"items/2x",
			"monsters":             				"monsters/2x",
			filepath.Join("ui", "alignments"):  "alignments/2x",
			filepath.Join("ui", "challenges"):  "challenges/2x",
			filepath.Join("ui", "companions"):  "companions/2x",
			filepath.Join("ui", "cosmetics"):   "cosmetics/2x",
			filepath.Join("ui", "emblems"):     "emblems/big",
			filepath.Join("ui", "emotes"):      "emotes/2x",
			filepath.Join("ui", "jobs"):        "jobs/2x",
			filepath.Join("ui", "mounts"):      "mounts/big",
			filepath.Join("ui", "presets"):     "presets/2x",
			filepath.Join("ui", "smilies"):     "smilies/2x",
			filepath.Join("ui", "spells"):      "spells/2x",
		}

		for key, path := range renamePaths {
			err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", path), filepath.Join(dir, "images", key))
			if err != nil {
				return err
			}
		}
		
		removePaths := []string{
			filepath.Join(outPath, "Assets"),
			filepath.Join(outPath, "monster_images.imagebundle"),
			filepath.Join(outPath, "mount_images.imagebundle"),
			filepath.Join(outPath, "spell_images.imagebundle"),
			filepath.Join(outPath, "spellstate_images.imagebundle"),
			filepath.Join(outPath, "job_images.imagebundle"),
			filepath.Join(outPath, "preset_images.imagebundle"),
			filepath.Join(outPath, "smiley_images.imagebundle"),
		}

		for _, path := range removePaths {
			err = os.RemoveAll(path)
			if err != nil {
				return err
			}
		}

		cleaningTasks := []struct {
			path    string
			dim     int
			exclude *regexp.Regexp
		}{
			{itemsPath, 128, nil},
			{monstersPath, 128, nil},
			{filepath.Join(uiPath, "achievements"), 58, nil},
			{filepath.Join(uiPath, "alignments"), 0, nil},
			{filepath.Join(uiPath, "arena"), 0, regexp.MustCompile(`(left|right|middle)BG`)},
			{filepath.Join(uiPath, "challenges"), 0, nil},
			{filepath.Join(uiPath, "companions"), 168, nil},
			{filepath.Join(uiPath, "cosmetics"), 128, nil},
			{filepath.Join(uiPath, "document"), 0, nil},
			{filepath.Join(uiPath, "emblems", "backcontent", "2x"), 0, nil},
			{filepath.Join(uiPath, "emblems", "outlinealliance", "2x"), 0, nil},
			{filepath.Join(uiPath, "emblems", "outlineguild", "2x"), 0, nil},
			{filepath.Join(uiPath, "emblems", "up", "2x"), 0, nil},
			{filepath.Join(uiPath, "emotes"), 0, nil},
			{filepath.Join(uiPath, "guidebook"), 0, nil},
			{filepath.Join(uiPath, "guildrank"), 0, nil},
			{filepath.Join(uiPath, "house"), 0, nil},
			{filepath.Join(uiPath, "icon"), 0, nil},
			{filepath.Join(uiPath, "illus"), 0, regexp.MustCompile(`^\d`)},
			{filepath.Join(uiPath, "jobs"), 0, nil},
			{filepath.Join(uiPath, "mounts"), 256, nil},
			{filepath.Join(uiPath, "ornament"), 0, nil},
			{filepath.Join(uiPath, "presets"), 96, nil},
			{filepath.Join(uiPath, "smilies"), 64, nil},
			{filepath.Join(uiPath, "spells"), 0, nil},
			{filepath.Join(uiPath, "spellstates"), 0, nil},
			{filepath.Join(uiPath, "suggestion"), 200, nil},
		}

		for _, task := range cleaningTasks {
			err = cleanImages(task.path, task.dim, task.exclude)
			if err != nil {
				return err
			}
		}

		emblemPaths := []string{
			filepath.Join(uiPath, "emblems", "backcontent", "2x"),
			filepath.Join(uiPath, "emblems", "outlinealliance", "2x"),
			filepath.Join(uiPath, "emblems", "outlineguild", "2x"),
			filepath.Join(uiPath, "emblems", "up", "2x"),
		}

		for _, path := range emblemPaths {
			err = moveFilesToParentFolder(path)
			if err != nil {
				return err
			}
		}

		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
