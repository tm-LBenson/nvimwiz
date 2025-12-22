local M = {}

function M.spec()
	return {
		{
			"catppuccin/nvim",
			name = "catppuccin",
			priority = 1000,
			config = function()
				require("catppuccin").setup({})
				vim.cmd("colorscheme catppuccin")
			end,
		},
	}
end

return M
