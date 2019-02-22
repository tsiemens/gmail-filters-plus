package cmd

import (
	gm "google.golang.org/api/gmail/v1"

	"github.com/spf13/cobra"
	"github.com/tsiemens/gmail-tools/api"
	"github.com/tsiemens/gmail-tools/config"
	"github.com/tsiemens/gmail-tools/prnt"
	"github.com/tsiemens/gmail-tools/util"
)

var searchLabels []string
var searchLabelsToAdd []string
var searchTouch = false
var searchTrash = false
var searchInteresting = false
var searchUninteresting = false
var searchPrintIdsOnly = false
var searchPrintJson = false
var searchMaxMsgs int64

func runSearchCmd(cmd *cobra.Command, args []string) {
	if searchInteresting && searchUninteresting {
		prnt.StderrLog.Fatalln("-u and -i options are mutually exclusive")
	}

	query := ""
	for _, label := range searchLabels {
		query += "label:(" + label + ") "
	}
	if len(args) > 0 {
		query += args[0] + " "
	}

	if query == "" {
		prnt.StderrLog.Println("No query provided")
	}

	conf := config.AppConfig()
	if searchTouch && conf.ApplyLabelOnTouch == "" {
		prnt.StderrLog.Fatalf("No ApplyLabelOnTouch property found in %s\n",
			conf.ConfigFile)
	}

	srv := api.NewGmailClient(api.ModifyScope)
	gHelper := NewGmailHelper(srv, api.DefaultUser, conf)

	msgs, err := gHelper.Msgs.QueryMessages(query, false, false, searchMaxMsgs, api.IdsOnly)
	if err != nil {
		prnt.StderrLog.Fatalf("%v\n", err)
	}
	prnt.LPrintf(prnt.Debug, "Debug: Query returned %d mesages\n", len(msgs))

	hasLoadedMsgDetails := false

	if searchInteresting || searchUninteresting {
		var filteredMsgs []*gm.Message
		msgs, err = gHelper.Msgs.LoadMessages(msgs, api.MessageFormatMetadata)
		util.CheckErr(err)
		for _, msg := range msgs {
			msgInterest := gHelper.MsgInterest(msg)
			if (searchInteresting && msgInterest == Interesting) ||
				(searchUninteresting && msgInterest == Uninteresting) {
				filteredMsgs = append(filteredMsgs, msg)
			}
		}
		hasLoadedMsgDetails = true
		msgs = filteredMsgs
	}

	if len(msgs) == 0 {
		prnt.HPrintln(prnt.Always, "Query matched no messages")
		return
	}
	prnt.HPrintf(prnt.Always, "Query matched %d messages\n", len(msgs))

	if !Quiet && MaybeConfirmFromInput("Show messages?", true) {
		if searchPrintIdsOnly {
			for _, msg := range msgs {
				prnt.Printf("%s,%s\n", msg.Id, msg.ThreadId)
			}
		} else {
			if !hasLoadedMsgDetails {
				msgs, err = gHelper.Msgs.LoadMessages(msgs, api.MessageFormatMetadata)
				util.CheckErr(err)
				hasLoadedMsgDetails = true
			}

			if searchPrintJson {
				gHelper.PrintMessagesJson(msgs)
			} else {
				gHelper.PrintMessagesByCategory(msgs)
			}
		}
	}

	if len(searchLabelsToAdd) > 0 {
		maybeApplyLabels(msgs, gHelper, searchLabelsToAdd)
	}
	if searchTouch {
		maybeTouchMessages(msgs, gHelper)
	}
	if searchTrash {
		maybeTrashMessages(msgs, gHelper)
	}
}

var searchCmd = &cobra.Command{
	Use:     "search [QUERY]",
	Short:   "Searches for messages with the given query",
	Aliases: []string{"find"},
	Run:     runSearchCmd,
	Args:    cobra.RangeArgs(0, 1),
}

func init() {
	RootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringArrayVarP(&searchLabels, "labelp", "l", []string{},
		"Label regexps to match in the search (may be provided multiple times)")
	searchCmd.Flags().StringArrayVar(&searchLabelsToAdd, "add-label", []string{},
		"Apply a label to matches (may be provided multiple times)")
	searchCmd.Flags().BoolVarP(&searchTouch, "touch", "t", false,
		"Apply 'touched' label from ~/.gmailcli/config.yaml")
	searchCmd.Flags().BoolVar(&searchTrash, "trash", false,
		"Send messages to the trash")
	searchCmd.Flags().BoolVarP(&searchInteresting, "interesting", "i", false,
		"Filter results by interesting messages")
	searchCmd.Flags().BoolVarP(&searchUninteresting, "uninteresting", "u", false,
		"Filter results by uninteresting messages")
	searchCmd.Flags().BoolVar(&searchPrintIdsOnly, "ids-only", false,
		"Only prints out only messageId,threadId (does not prompt)")
	searchCmd.Flags().BoolVar(&searchPrintJson, "json", false,
		"Print message details formatted as json")
	searchCmd.Flags().Int64VarP(&searchMaxMsgs, "max", "m", -1,
		"Set a max on how many results are queried.")
	addDryFlag(searchCmd)
}
