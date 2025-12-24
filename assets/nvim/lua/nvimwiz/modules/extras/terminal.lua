local M = {}

function M.spec()
  return {
    {
      "akinsho/toggleterm.nvim",
      version = "*",
      config = function()
        local ok, toggleterm = pcall(require, "toggleterm")
        if not ok then
          return
        end

        toggleterm.setup({
          direction = "horizontal",
          size = 15,
          start_in_insert = true,
          insert_mappings = true,
          persist_size = true,
          shade_terminals = true,
        })

        -- Toggle a bottom terminal split.
        vim.keymap.set({ "n", "t" }, "<leader>tt", "<cmd>ToggleTerm direction=horizontal<cr>", { desc = "Toggle terminal" })

        -- Exit terminal mode (use double-esc to avoid breaking terminal apps).
        vim.keymap.set("t", "<esc><esc>", [[<C-\\><C-n>]], { desc = "Exit terminal mode" })
      end,
    },
  }
end

return M
