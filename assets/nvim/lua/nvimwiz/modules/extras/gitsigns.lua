local M = {}

function M.spec()
	return {
		{
			"lewis6991/gitsigns.nvim",
			config = function()
				require("gitsigns").setup({})
			end,
		},
	}
end

return M
