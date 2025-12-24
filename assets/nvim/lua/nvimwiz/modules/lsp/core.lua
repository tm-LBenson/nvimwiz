local M = {}

local function uniq(list)
  local seen = {}
  local out = {}
  for _, v in ipairs(list) do
    if v and v ~= "" and not seen[v] then
      seen[v] = true
      table.insert(out, v)
    end
  end
  return out
end

local function has_config(configs, name)
  return configs and configs[name] ~= nil
end

local function add_one(configs, servers, name)
  if not has_config(configs, name) then
    return false
  end
  table.insert(servers, name)
  return true
end

local function add_prefer(configs, servers, preferred, fallback)
  if add_one(configs, servers, preferred) then
    return preferred
  end
  if fallback and add_one(configs, servers, fallback) then
    return fallback
  end
  return nil
end

local function merge_capabilities(base, extra)
  if not extra then
    return base
  end
  return vim.tbl_deep_extend("force", {}, base, extra)
end

function M.spec()
  return {
    {
      "neovim/nvim-lspconfig",
      dependencies = { "williamboman/mason.nvim", "williamboman/mason-lspconfig.nvim" },
      config = function()
        local cfg = require("nvimwiz.generated.config")
        local wants = (cfg and cfg.lsp) or {}

        require("mason").setup({})
        local mlsp = require("mason-lspconfig")

        local configs = nil
        pcall(function()
          configs = require("lspconfig.configs")
        end)

        local servers = {}

        if wants.typescript then
          add_prefer(configs, servers, "ts_ls", "tsserver")
        end
        if wants.python then
          add_one(configs, servers, "pyright")
        end
        if wants.web then
          add_one(configs, servers, "html")
          add_one(configs, servers, "cssls")
        end
        if wants.emmet then
          add_prefer(configs, servers, "emmet_ls", "emmet_language_server")
        end
        if wants.go then
          add_one(configs, servers, "gopls")
        end
        if wants.bash then
          add_one(configs, servers, "bashls")
        end
        if wants.lua then
          add_one(configs, servers, "lua_ls")
        end
        if wants.java then
          add_one(configs, servers, "jdtls")
        end

        servers = uniq(servers)

        mlsp.setup({ ensure_installed = servers, automatic_installation = true })

        local caps = require("nvimwiz.lsp").capabilities()

        local function setup_server(server_name)
          local opts = {}

          if server_name == "lua_ls" then
            opts.settings = {
              Lua = {
                diagnostics = { globals = { "vim" } },
                workspace = { checkThirdParty = false },
                telemetry = { enable = false },
              },
            }
          end

          if server_name == "emmet_ls" or server_name == "emmet_language_server" then
            opts.filetypes = {
              "html",
              "css",
              "scss",
              "less",
              "javascript",
              "javascriptreact",
              "typescript",
              "typescriptreact",
              "svelte",
              "vue",
            }
          end

          opts.capabilities = merge_capabilities(caps, opts.capabilities)

          if vim.lsp and vim.lsp.enable and vim.lsp.config then
            pcall(function()
              vim.lsp.config(server_name, opts)
            end)
            pcall(function()
              vim.lsp.enable(server_name)
            end)
            return
          end

          local ok, lspconfig = pcall(require, "lspconfig")
          if not ok then
            return
          end
          local server = lspconfig[server_name]
          if server and server.setup then
            pcall(server.setup, opts)
          end
        end

        for _, server_name in ipairs(servers) do
          setup_server(server_name)
        end

        vim.api.nvim_create_autocmd("LspAttach", {
          callback = function(args)
            local buf = args.buf
            vim.keymap.set("n", "K", vim.lsp.buf.hover, { buffer = buf, desc = "LSP Hover" })
            vim.keymap.set("n", "gd", vim.lsp.buf.definition, { buffer = buf, desc = "LSP Definition" })
            vim.keymap.set("n", "gr", vim.lsp.buf.references, { buffer = buf, desc = "LSP References" })
            vim.keymap.set("n", "<leader>rn", vim.lsp.buf.rename, { buffer = buf, desc = "LSP Rename" })
          end,
        })
      end,
    },
  }
end

return M
