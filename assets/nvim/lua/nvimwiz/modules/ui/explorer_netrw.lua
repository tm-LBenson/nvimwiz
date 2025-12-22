local M = {}

function M.setup()
	vim.g.netrw_banner = 0
	vim.g.netrw_liststyle = 3
	vim.g.netrw_browse_split = 0
	vim.g.netrw_winsize = 25
	vim.keymap.set("n", "<leader>e", "<cmd>Explore<CR>", { desc = "Explorer" })
end

return M
