local M = {}

function M.spec()
	return {
		{
			"folke/which-key.nvim",
			config = function()
				require("which-key").setup({})
			end,
		},
	}
end

return M
