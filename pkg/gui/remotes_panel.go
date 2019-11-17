package gui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands"
	"github.com/jesseduffield/lazygit/pkg/utils"
)

// list panel functions

func (gui *Gui) getSelectedRemote() *commands.Remote {
	selectedLine := gui.State.Panels.Remotes.SelectedLine
	if selectedLine == -1 {
		return nil
	}

	return gui.State.Remotes[selectedLine]
}

func (gui *Gui) handleRemotesClick(g *gocui.Gui, v *gocui.View) error {
	itemCount := len(gui.State.Remotes)
	handleSelect := gui.handleRemoteSelect
	selectedLine := &gui.State.Panels.Remotes.SelectedLine

	return gui.handleClick(v, itemCount, selectedLine, handleSelect)
}

func (gui *Gui) handleRemoteSelect(g *gocui.Gui, v *gocui.View) error {
	if gui.popupPanelFocused() {
		return nil
	}

	gui.State.SplitMainPanel = false

	if _, err := gui.g.SetCurrentView(v.Name()); err != nil {
		return err
	}

	gui.getMainView().Title = "Remote"

	remote := gui.getSelectedRemote()
	if err := gui.focusPoint(0, gui.State.Panels.Remotes.SelectedLine, len(gui.State.Remotes), v); err != nil {
		return err
	}

	return gui.renderString(g, "main", fmt.Sprintf("%s\nUrls:\n%s", utils.ColoredString(remote.Name, color.FgGreen), strings.Join(remote.Urls, "\n")))
}

// gui.refreshStatus is called at the end of this because that's when we can
// be sure there is a state.Remotes array to pick the current remote from
func (gui *Gui) refreshRemotes() error {
	remotes, err := gui.GitCommand.GetRemotes()
	if err != nil {
		return gui.createErrorPanel(gui.g, err.Error())
	}

	gui.State.Remotes = remotes

	if gui.getBranchesView().Context == "remotes" {
		return gui.renderRemotesWithSelection()
	}

	return nil
}

func (gui *Gui) renderRemotesWithSelection() error {
	branchesView := gui.getBranchesView()

	gui.refreshSelectedLine(&gui.State.Panels.Remotes.SelectedLine, len(gui.State.Remotes))
	if err := gui.renderListPanel(branchesView, gui.State.Remotes); err != nil {
		return err
	}
	if err := gui.handleRemoteSelect(gui.g, branchesView); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) handleRemoteEnter(g *gocui.Gui, v *gocui.View) error {
	// naive implementation: get the branches and render them to the list, change the context
	remote := gui.getSelectedRemote()

	gui.State.RemoteBranches = remote.Branches

	newSelectedLine := 0
	if len(remote.Branches) == 0 {
		newSelectedLine = -1
	}
	gui.State.Panels.RemoteBranches.SelectedLine = newSelectedLine

	return gui.switchBranchesPanelContext("remote-branches")
}

func (gui *Gui) handleAddRemote(g *gocui.Gui, v *gocui.View) error {
	branchesView := gui.getBranchesView()
	return gui.createPromptPanel(g, branchesView, gui.Tr.SLocalize("newRemoteName"), "", func(g *gocui.Gui, v *gocui.View) error {
		remoteName := gui.trimmedContent(v)
		return gui.createPromptPanel(g, branchesView, gui.Tr.SLocalize("newRemoteUrl"), "", func(g *gocui.Gui, v *gocui.View) error {
			remoteUrl := gui.trimmedContent(v)
			if err := gui.GitCommand.AddRemote(remoteName, remoteUrl); err != nil {
				return err
			}
			return gui.refreshRemotes()
		})
	})
}

func (gui *Gui) handleRemoveRemote(g *gocui.Gui, v *gocui.View) error {
	remote := gui.getSelectedRemote()
	if remote == nil {
		return nil
	}
	return gui.createConfirmationPanel(g, v, true, gui.Tr.SLocalize("removeRemote"), gui.Tr.SLocalize("removeRemotePrompt")+" '"+remote.Name+"'?", func(*gocui.Gui, *gocui.View) error {
		if err := gui.GitCommand.RemoveRemote(remote.Name); err != nil {
			return err
		}

		return gui.refreshRemotes()

	}, nil)
}