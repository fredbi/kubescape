package v2

import (
	"log"
	"sort"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

// ensureTableLayout performs a best effort to render the table nicely on a terminal.
func ensureTableLayout(table *tablewriter.Table, cols int, lines [][]string) {
	const defaultTermWidth = 132

	// determine if we need to wrap columns to adjust to the terminal
	isTerminal := term.IsTerminal(0)
	if !isTerminal {
		table.SetAutoWrapText(false)

		return
	}

	width, _, err := term.GetSize(0)
	if err != nil {
		width = defaultTermWidth
	}
	width -= cols + 1 // minus borders

	colWidth := computeColumnsWidth(width, cols, lines) // heuristic to determine the least constraining max column width
	if colWidth > 0 {
		table.SetColWidth(colWidth) // max col width
		table.SetAutoWrapText(true)
	} else {
		table.SetAutoWrapText(false)
	}

	log.Println("DEBUG(FRED)", isTerminal, width, colWidth, cols)
}

// computeColumnsWidth determines the max width to apply to all columns, given the terminal width
// and all the lines to be displayed (including headings).
//
// NOTE: this feature could be a nice addition to github.com/olekukonko/tablewriter,
// as that package currently supports a similar feature, but for minimum width.
//
// i)  compute the fixed-bucket histogram of display widths per column (p-values)
// ii) iterate over deciles, from 90% down to 10%
//       *iterate over columns, widest first:
//         * replace max width for column at decile:  if new capped total width =< available, max length = this value
// iii) WrapString() with this new spec
func computeColumnsWidth(available, cols int, lines [][]string) int {
	type colWidth struct {
		ColIndex int
		MaxWidth int
	}

	maxWidths := make(map[int]int, cols)
	widths := make([][]int, cols)
	orderedMaxWidths := make([]colWidth, 0, cols)

	// assess the histogram of widths for each column.
	for _, line := range lines {
		for j, column := range line {
			colWidth := tablewriter.DisplayWidth(column)
			widths[j] = append(widths[j], colWidth)
			currentColWidth := maxWidths[j]
			if colWidth > currentColWidth {
				maxWidths[j] = colWidth
			}
		}
	}
	for j := range widths {
		sort.Ints(widths[j])
	}

	// determine total width : compare with the available width
	var total int
	for idx, width := range maxWidths {
		total += width
		orderedMaxWidths = append(orderedMaxWidths, colWidth{ColIndex: idx, MaxWidth: width})
	}

	if total <= available {
		return -1 // no need to bother with wrapping columns: everything fits in
	}

	// order columns by decreasing max width
	sort.Slice(orderedMaxWidths, func(i, j int) bool {
		return orderedMaxWidths[i].MaxWidth > orderedMaxWidths[j].MaxWidth
	})

	if available < orderedMaxWidths[len(orderedMaxWidths)-1].MaxWidth {
		return -1 // no need to bother with wrapping columns: there is nothing we can do and the output will not look nice
	}

	return available / cols

	/*
		// determine columns by order of requested width
		//_, maxWidth := largestInMap(maxWidths)

		// gap := (total - available)

		return 0 // TODO
	*/
}

/*
func largestInMap(in map[int]int) (int, int) {
	var largest int
	index := -1
	for k, v := range in {
		if v > largest || index == -1 {
			index = k
			largest = v
		}
	}

	return index, largest
}

func smallestInMap(in map[int]int) (int, int) {
	var smallest int
	index := -1
	for k, v := range in {
		if v < smallest || index == -1 {
			index = k
			smallest = v
		}
	}

	return index, smallest
}
*/
