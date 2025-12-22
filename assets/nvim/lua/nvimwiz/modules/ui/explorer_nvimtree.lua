local M = {}

function M.spec()
	return {
		{
			"nvim-tree/nvim-tree.lua",
			dependencies = { "nvim-tree/nvim-web-devicons" },
			config = function()
				require("nvim-tree").setup({
					filters = { dotfiles = false },
					view = { width = 35 },
					renderer = { group_empty = true },
					update_focused_file = { enable = true, update_root = true },
				})
				vim.keymap.set("n", "<leader>e", "<cmd>NvimTreeToggle<CR>", { desc = "Explorer" })
			end,
		},
		{ "nvim-tree/nvim-web-devicons", lazy = true },
	}
end

return M
