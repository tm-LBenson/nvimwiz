local M = {}

function M.spec()
	return {
		{
			"nvim-telescope/telescope.nvim",
			dependencies = { "nvim-lua/plenary.nvim" },
			config = function()
				require("telescope").setup({})
				local b = require("telescope.builtin")
				vim.keymap.set("n", "<leader>ff", b.find_files, { desc = "Find files" })
				vim.keymap.set("n", "<leader>fg", b.live_grep, { desc = "Live grep" })
				vim.keymap.set("n", "<leader>fb", b.buffers, { desc = "Buffers" })
				vim.keymap.set("n", "<leader>fh", b.help_tags, { desc = "Help" })
			end,
		},
	}
end

return M
