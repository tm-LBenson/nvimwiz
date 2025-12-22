local M = {}

local function add_if_exists(configs, servers, name)
	if configs[name] then
		table.insert(servers, name)
		return true
	end
	return false
end

function M.spec()
	return {
		{
			"williamboman/mason.nvim",
			config = function()
				require("mason").setup({})
			end,
		},
		{
			"williamboman/mason-lspconfig.nvim",
			dependencies = { "williamboman/mason.nvim", "neovim/nvim-lspconfig" },
			config = function()
				local cfg = require("nvimwiz.util").config()
				local wants = cfg.lsp or {}
				local configs = require("lspconfig.configs")
				local servers = {}

				if wants.typescript then
					if not add_if_exists(configs, servers, "tsserver") then
						add_if_exists(configs, servers, "ts_ls")
					end
				end
				if wants.python then
					add_if_exists(configs, servers, "pyright")
				end
				if wants.web then
					add_if_exists(configs, servers, "html")
					add_if_exists(configs, servers, "cssls")
				end
				if wants.go then
					add_if_exists(configs, servers, "gopls")
				end
				if wants.bash then
					add_if_exists(configs, servers, "bashls")
				end
				if wants.lua then
					add_if_exists(configs, servers, "lua_ls")
				end
				if wants.java then
					add_if_exists(configs, servers, "jdtls")
				end

				require("mason-lspconfig").setup({
					ensure_installed = servers,
					automatic_installation = true,
				})

				local lspconfig = require("lspconfig")
				local lsp = require("nvimwiz.lsp")
				local capabilities = lsp.capabilities()

				vim.diagnostic.config({
					virtual_text = true,
					signs = true,
					underline = true,
					update_in_insert = false,
					severity_sort = true,
				})

				require("mason-lspconfig").setup_handlers({
					function(server)
						if server == "jdtls" then
							return
						end
						local opts = { on_attach = lsp.on_attach, capabilities = capabilities }
						if server == "lua_ls" then
							opts.settings = {
								Lua = {
									diagnostics = { globals = { "vim" } },
								},
							}
						end
						lspconfig[server].setup(opts)
					end,
				})
			end,
		},
		{ "neovim/nvim-lspconfig" },
	}
end

return M
