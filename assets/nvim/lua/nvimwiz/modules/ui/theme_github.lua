local M = {}

function M.spec()
  return {
    {
      "projekt0n/github-nvim-theme",
      name = "github-theme",
      lazy = false,
      priority = 1000,
      config = function()
        -- GitHub theme provides multiple variants:
        -- github_dark, github_dark_dimmed, github_dark_high_contrast, github_light, etc.
        require("github-theme").setup({})
        vim.cmd("colorscheme github_dark")
      end,
    },
  }
end

return M
