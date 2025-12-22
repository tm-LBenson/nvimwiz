local M = {}

function M.spec()
	return {
		{
			"windwp/nvim-autopairs",
			config = function()
				require("nvim-autopairs").setup({})
			end,
		},
	}
end

return M
