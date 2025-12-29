local ok_cfg, cfg = pcall(require, "nvimwiz.generated.config")
if not ok_cfg or type(cfg) ~= "table" then
	cfg = {}
end

vim.g.mapleader = cfg.leader or " "
vim.g.maplocalleader = cfg.localleader or " "

vim.opt.termguicolors = true

local ln = "relative"
if type(cfg.choices) == "table" and type(cfg.choices.line_numbers) == "string" then
	ln = cfg.choices.line_numbers
end

if ln == "absolute" then
	vim.opt.number = true
	vim.opt.relativenumber = false
elseif ln == "off" then
	vim.opt.number = false
	vim.opt.relativenumber = false
else
	vim.opt.number = true
	vim.opt.relativenumber = true
end
vim.opt.expandtab = true
vim.opt.shiftwidth = 2
vim.opt.tabstop = 2
vim.opt.smartindent = true
vim.opt.signcolumn = "yes"
vim.opt.updatetime = 250
vim.opt.timeoutlen = 400

local uv = vim.uv or vim.loop
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not uv.fs_stat(lazypath) then
	vim.fn.system({ "git", "clone", "--filter=blob:none", "https://github.com/folke/lazy.nvim.git", "--branch=stable", lazypath })
end
vim.opt.rtp:prepend(lazypath)

local specs = {}
local setups = {}

for _, modname in ipairs(cfg.modules or {}) do
	local ok, m = pcall(require, modname)
	if ok and type(m) == "table" then
		if type(m.spec) == "function" then
			local s = m.spec()
			if type(s) == "table" then
				for _, v in ipairs(s) do
					table.insert(specs, v)
				end
			end
		end
		if type(m.setup) == "function" then
			table.insert(setups, m.setup)
		end
	end
end

require("lazy").setup(specs, {
	defaults = { lazy = false },
	checker = { enabled = false },
	change_detection = { enabled = false },
	ui = { border = "rounded" },
})

for _, fn in ipairs(setups) do
	pcall(fn)
end

pcall(require, "nvimwiz.user")
