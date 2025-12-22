local M = {}

function M.spec()
	return {
		{
			"nvim-treesitter/nvim-treesitter",
			build = ":TSUpdate",
			config = function()
				require("nvim-treesitter.configs").setup({
					highlight = { enable = true },
					indent = { enable = true },
					ensure_installed = { "bash", "css", "go", "html", "java", "javascript", "lua", "python", "typescript", "vim", "vimdoc" },
				})
			end,
		},
	}
end

return M
