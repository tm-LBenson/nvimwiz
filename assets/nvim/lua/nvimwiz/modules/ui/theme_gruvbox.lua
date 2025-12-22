local M = {}

function M.spec()
	return {
		{
			"ellisonleao/gruvbox.nvim",
			priority = 1000,
			config = function()
				require("gruvbox").setup({})
				vim.cmd("colorscheme gruvbox")
			end,
		},
	}
end

return M
