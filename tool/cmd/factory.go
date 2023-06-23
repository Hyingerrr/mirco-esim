package cmd

import (
	filedir "github.com/Hyingerrr/mirco-esim/pkg/file-dir"
	"github.com/Hyingerrr/mirco-esim/pkg/templates"
	"github.com/Hyingerrr/mirco-esim/tool/factory"

	"github.com/spf13/cobra"
)

var factoryCmd = &cobra.Command{
	Use:   "factory",
	Short: "初始化结构体",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		esimFactory := factory.NewEsimFactory(
			factory.WithEsimFactoryLogger(logger),
			factory.WithEsimFactoryWriter(filedir.NewEsimWriter()),
			factory.WithEsimFactoryTpl(templates.NewTextTpl()),
		)
		err := esimFactory.Run(v)
		if err != nil {
			logger.Errorf(err.Error())
		}
		esimFactory.Close()
	},
}

func init() {
	rootCmd.AddCommand(factoryCmd)

	factoryCmd.Flags().BoolP("sort", "s", true, "sort the field")

	factoryCmd.Flags().BoolP("new", "n", false, "with new")

	factoryCmd.Flags().BoolP("option", "o", false, "New with option")

	factoryCmd.Flags().BoolP("pool", "p", false, "with pool")

	factoryCmd.Flags().BoolP("ol", "", false, "generate logger option")

	factoryCmd.Flags().BoolP("oc", "", false, "generate conf option")

	factoryCmd.Flags().BoolP("print", "", false, "print to terminal")

	factoryCmd.Flags().StringP("sname", "", "", "struct name")

	factoryCmd.Flags().StringP("sdir", "", "", "struct path")

	factoryCmd.Flags().BoolP("plural", "", false, "with plural")

	factoryCmd.Flags().StringP("imp_iface", "", "", "implement the interface")

	err := v.BindPFlags(factoryCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
