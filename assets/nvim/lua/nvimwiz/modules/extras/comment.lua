local M = {}

function M.spec()
	return {
		{
			"numToStr/Comment.nvim",
			config = function()
				require("Comment").setup({})
			end,
		},
	}
end

return M
