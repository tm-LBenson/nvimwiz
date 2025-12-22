local M = {}

local function native_api()
	return vim.lsp and vim.lsp.config ~= nil and vim.lsp.enable ~= nil
end

local function has_config(name)
	local p = "lsp/" .. name .. ".lua"
	local files = vim.api.nvim_get_runtime_file(p, false)
	return files and #files > 0
end

local function push_unique(list, v)
	if not v or v == "" then
		return
	end
	for _, x in ipairs(list) do
		if x == v then
			return
		end
	end
	list[#list + 1] = v
end

local function pick_ts(use_native)
	if use_native then
		if has_config("ts_ls") then
			return "ts_ls"
		end
		if has_config("tsserver") then
			return "tsserver"
		end
		return "ts_ls"
	end
	return "tsserver"
end

function M.spec()
	return {
		{
			"mason-org/mason.nvim",
			config = function()
				local ok, mason = pcall(require, "mason")
				if ok then
					mason.setup({})
				end
			end,
		},
		{
			"mason-org/mason-lspconfig.nvim",
			dependencies = { "mason-org/mason.nvim", "neovim/nvim-lspconfig" },
			config = function()
				local ok_util, util = pcall(require, "nvimwiz.util")
				if not ok_util then
					return
				end

				local cfg = util.config() or {}
				local wants = cfg.lsp or {}

				local use_native = native_api()

				local install = {}
				local enable = {}

				local function add(name, also_enable)
					if also_enable and use_native and not has_config(name) then
						push_unique(install, name)
						return
					end
					push_unique(install, name)
					if also_enable then
						push_unique(enable, name)
					end
				end

				if wants.typescript then
					add(pick_ts(use_native), true)
				end
				if wants.python then
					add("pyright", true)
				end
				if wants.web then
					add("html", true)
					add("cssls", true)
				end
				if wants.go then
					add("gopls", true)
				end
				if wants.bash then
					add("bashls", true)
				end
				if wants.lua then
					add("lua_ls", true)
				end
				if wants.java then
					add("jdtls", false)
				end

				local ok_mlsp, mlsp = pcall(require, "mason-lspconfig")
				if ok_mlsp then
					local opts = { ensure_installed = install }
					if type(mlsp.setup_handlers) == "function" then
						opts.automatic_installation = true
					else
						opts.automatic_enable = false
					end
					mlsp.setup(opts)
				end

				local ok_lsp, lsp = pcall(require, "nvimwiz.lsp")
				if not ok_lsp then
					return
				end

				local capabilities = lsp.capabilities()

				vim.diagnostic.config({
					virtual_text = true,
					signs = true,
					underline = true,
					update_in_insert = false,
					severity_sort = true,
				})

				local function mk_opts(server)
					local opts = { on_attach = lsp.on_attach, capabilities = capabilities }
					if server == "lua_ls" then
						opts.settings = { Lua = { diagnostics = { globals = { "vim" } } } }
					end
					return opts
				end

				if use_native then
					for _, server in ipairs(enable) do
						pcall(vim.lsp.config, server, mk_opts(server))
					end
					if #enable > 0 then
						pcall(vim.lsp.enable, enable)
					end
					return
				end

				local ok_lspconfig, lspconfig = pcall(require, "lspconfig")
				if not ok_lspconfig then
					return
				end

				for _, server in ipairs(enable) do
					if lspconfig[server] and type(lspconfig[server].setup) == "function" then
						lspconfig[server].setup(mk_opts(server))
					end
				end
			end,
		},
		{ "neovim/nvim-lspconfig" },
	}
end

return M
