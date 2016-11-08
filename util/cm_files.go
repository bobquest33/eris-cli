package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"

	"github.com/BurntSushi/toml"
)

// XXX: this is temporary until eris-keys.js is more tightly integrated with eris-contracts.js
type accountInfo struct {
	Address string `mapstructure:"address" json:"address" yaml:"address" toml:"address"`
	PubKey  string `mapstructure:"pubKey" json:"pubKey" yaml:"pubKey" toml:"pubKey"`
	PrivKey string `mapstructure:"privKey" json:"privKey" yaml:"privKey" toml:"privKey"`
}

func SaveAccountResults(do *definitions.Do) error {
	addrFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "addresses.csv"))
	if err != nil {
		return fmt.Errorf("Error creating addresses file. This usually means that there was a problem with the chain making process.")
	}
	defer addrFile.Close()

	log.WithField("name", do.Name).Debug("Creating file")
	actFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "accounts.csv"))
	if err != nil {
		return fmt.Errorf("Error creating accounts file.")
	}
	log.WithField("path", filepath.Join(config.ChainsPath, do.Name, "accounts.csv")).Debug("File successfully created")
	defer actFile.Close()

	log.WithField("name", do.Name).Debug("Creating file")
	actJSONFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "accounts.json"))
	if err != nil {
		return fmt.Errorf("Error creating accounts file.")
	}
	log.WithField("path", filepath.Join(config.ChainsPath, do.Name, "accounts.json")).Debug("File successfully created")
	defer actJSONFile.Close()

	valFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "validators.csv"))
	if err != nil {
		return fmt.Errorf("Error creating validators file.")
	}
	defer valFile.Close()

	accountJsons := make(map[string]*accountInfo)

	for _, account := range do.Accounts {
		accountJsons[account.Name] = &accountInfo{
			Address: account.Address,
			PubKey:  account.PubKey,
			PrivKey: account.MintKey.PrivKey[1].(string),
		}

		_, err := addrFile.WriteString(fmt.Sprintf("%s,%s\n", account.Address, account.Name))
		if err != nil {
			log.Error("Error writing addresses file.")
			return err
		}
		_, err = actFile.WriteString(fmt.Sprintf("%s,%d,%s,%d,%d\n", account.PubKey, account.Tokens, account.Name, account.MintPermissions.MintBase.MintPerms, account.MintPermissions.MintBase.MintSetBit))
		if err != nil {
			log.Error("Error writing accounts file.")
			return err
		}
		if account.Validator {
			_, err = valFile.WriteString(fmt.Sprintf("%s,%d,%s,%d,%d\n", account.PubKey, account.ToBond, account.Name, account.MintPermissions.MintBase.MintPerms, account.MintPermissions.MintBase.MintSetBit))
			if err != nil {
				log.Error("Error writing validators file.")
				return err
			}
		}
	}
	addrFile.Sync()
	actFile.Sync()
	valFile.Sync()

	j, err := json.MarshalIndent(accountJsons, "", "  ")
	if err != nil {
		return err
	}

	_, err = actJSONFile.Write(j)
	if err != nil {
		return err
	}

	log.WithField("path", actJSONFile.Name()).Debug("Saving File.")
	log.WithField("path", addrFile.Name()).Debug("Saving File.")
	log.WithField("path", actFile.Name()).Debug("Saving File.")
	log.WithField("path", valFile.Name()).Debug("Saving File.")

	return nil
}

// returns a list of filenames which are the account_types files
// these *should be* absolute paths, but this is not a contract
// with calling functions.
func AccountTypes(erisPath string) ([]string, error) {
	haveTyps, err := filepath.Glob(filepath.Join(erisPath, "*.toml"))
	if err != nil {
		return []string{}, err
	}
	return haveTyps, nil
}

func AccountTypesNames(erisPath string, withExt bool) ([]string, error) {
	files, err := AccountTypes(erisPath)
	if err != nil {
		return []string{}, err
	}
	names := []string{}
	for _, file := range files {
		names = append(names, filepath.Base(file))
	}
	if !withExt {
		for e, name := range names {
			names[e] = strings.Replace(name, ".toml", "", 1)
		}
	}
	return names, nil
}

func WriteGenesisFile(name string, genesis *definitions.MintGenesis, account *definitions.Account, single bool) error {
	return writer(genesis, name, account.Name, "genesis.json", single)
}

func WritePrivVals(name string, account *definitions.Account, single bool) error {
	return writer(account.MintKey, name, account.Name, "priv_validator.json", single)
}

func WriteConfigurationFile(chain_name, account_name, seeds string, single bool,
	chainImageName string, useDataContainer bool, exportedPorts []string, containerEntrypoint string) error {
	if account_name == "" {
		account_name = "anonymous_marmot"
	}
	if chain_name == "" {
		return fmt.Errorf("No chain name provided.")
	}
	var fileBytes []byte
	var err error
	if fileBytes, err = config.GetConfigurationFileBytes(chain_name,
		account_name, seeds, chainImageName, useDataContainer,
		convertExportPortsSliceToString(exportedPorts), containerEntrypoint); err != nil {
		return err
	}
	var file string
	if !single {
		file = filepath.Join(config.ChainsPath, chain_name, account_name, "config.toml")
	} else {
		file = filepath.Join(config.ChainsPath, chain_name, "config.toml")
	}
	log.WithField("path", file).Debug("Saving File.")
	if err := config.WriteFile(string(fileBytes), file); err != nil {
		return err
	}
	return nil
}

func SaveAccountType(thisActT *definitions.AccountType) error {
	writer, err := os.Create(filepath.Join(config.AccountsTypePath, fmt.Sprintf("%s.toml", thisActT.Name)))
	defer writer.Close()
	if err != nil {
		return err
	}

	enc := toml.NewEncoder(writer)
	enc.Indent = ""
	err = enc.Encode(thisActT)
	if err != nil {
		return err
	}
	return nil
}

func convertExportPortsSliceToString(exportPorts []string) string {
	if len(exportPorts) == 0 {
		return ""
	}
	return `[ "` + strings.Join(exportPorts[:], `", "`) + `" ]`
}

func writer(toWrangle interface{}, chain_name, account_name, fileBase string, single bool) error {
	var file string
	fileBytes, err := json.MarshalIndent(toWrangle, "", "  ")
	if err != nil {
		return err
	}
	if !single {
		file = filepath.Join(config.ChainsPath, chain_name, account_name, fileBase)
	} else {
		file = filepath.Join(config.ChainsPath, chain_name, fileBase)
	}
	log.WithField("path", file).Debug("Saving File.")
	err = config.WriteFile(string(fileBytes), file)
	if err != nil {
		return err
	}
	return nil
}
