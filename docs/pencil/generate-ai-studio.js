const fs = require('fs');

function genId() {
  return Math.random().toString(36).substring(2, 7);
}

const aiStudioPage = {
  version: "2.8",
  children: [
    {
      type: "frame",
      id: genId(),
      x: 0,
      y: 0,
      name: "AI Movie Studio Page",
      width: 1440,
      height: 900,
      fill: "#000000",
      layout: "none",
      children: [
        // Background Image
        {
          type: "rectangle",
          id: genId(),
          x: 0,
          y: 0,
          name: "Background",
          fill: {
            type: "image",
            enabled: true,
            url: "./images/generated-1772088909104.png", // Reusing image from context
            mode: "fill"
          },
          width: 1440,
          height: 900
        },
        
        // Top Header
        {
          type: "frame",
          id: genId(),
          x: 0,
          y: 0,
          name: "Top Header",
          width: 1440,
          height: 80,
          padding: 24,
          fill: "transparent",
          layout: "horizontal",
          gap: "auto", // To push items to sides if supported, otherwise just use x/y for items inside none layout
          children: []
        },
        // We'll place Top Header items manually using absolute positioning to ensure it looks like the photo
        {
          type: "text",
          id: genId(),
          x: 40,
          y: 30,
          name: "Wondershare Logo",
          fill: "#FFFFFF",
          content: "wondershare\n万兴科技",
          fontFamily: "Inter",
          fontSize: 14,
          fontWeight: "bold"
        },
        // Right header container
        {
          type: "frame",
          id: genId(),
          x: 950,
          y: 20,
          name: "Right Actions",
          layout: "horizontal",
          gap: 16,
          height: 40,
          fill: "transparent",
          children: [
            {
              type: "frame",
              id: genId(),
              name: "Help Button",
              padding: 12,
              fill: "#FFFFFF1A",
              cornerRadius: 20,
              children: [
                { type: "text", id: genId(), fill: "#FFFFFF", content: "创作手册", fontSize: 14 }
              ]
            },
            {
              type: "frame",
              id: genId(),
              name: "Cooperation Button",
              padding: 12,
              fill: "#FFFFFF1A",
              cornerRadius: 20,
              children: [
                { type: "text", id: genId(), fill: "#FFFFFF", content: "商务合作", fontSize: 14 }
              ]
            },
            {
              type: "frame",
              id: genId(),
              name: "Credits",
              padding: 12,
              fill: "#FFFFFF1A",
              cornerRadius: 20,
              children: [
                { type: "text", id: genId(), fill: "#FFFFFF", content: "✨ 0", fontSize: 14 }
              ]
            },
            // Avatar Placeholder
            {
              type: "rectangle",
              id: genId(),
              name: "Avatar",
              width: 32,
              height: 32,
              cornerRadius: 16,
              fill: "#9C42F5"
            }
          ]
        },

        // Floating Left Sidebar
        {
          type: "frame",
          id: genId(),
          x: 24,
          y: 350,
          name: "Left Sidebar",
          width: 72,
          fill: "#2C2C2CB3",
          cornerRadius: 16,
          layout: "vertical",
          gap: 24,
          padding: 24,
          children: [
            { type: "text", id: genId(), name: "Home", fill: "#FFFFFF", content: "🏠\n首页", textAlign: "center", fontSize: 12 },
            { type: "text", id: genId(), name: "Script", fill: "#FFFFFF99", content: "📝\n剧本", textAlign: "center", fontSize: 12 },
            { type: "text", id: genId(), name: "Project", fill: "#FFFFFF99", content: "📁\n项目", textAlign: "center", fontSize: 12 },
            { type: "text", id: genId(), name: "Tools", fill: "#FFFFFF99", content: "🛠\n工具箱", textAlign: "center", fontSize: 12 },
            { type: "text", id: genId(), name: "Team", fill: "#FFFFFF99", content: "👥\n团队", textAlign: "center", fontSize: 12 }
          ]
        },

        // Center Content Block
        {
          type: "frame",
          id: genId(),
          x: 200,
          y: 200,
          name: "Center Content",
          width: 1040,
          height: 500,
          layout: "vertical",
          gap: 32,
          padding: 40,
          fill: "transparent",
          // The image has a logo, massive title, subtitle, and primary button neatly centered.
          children: [
            // Center Logo Row
            {
              type: "frame",
              id: genId(),
              name: "Logo Container",
              layout: "horizontal",
              width: "fill_container", // Need to center somehow, if no justify-content, maybe rely on text alignment or fixed widths.
              gap: 16,
              fill: "transparent",
              children: [
                {
                  type: "rectangle",
                  id: genId(),
                  width: 40,
                  height: 40,
                  fill: "#9C42F5", // the N logo placeholder
                  cornerRadius: 8
                },
                {
                  type: "text",
                  id: genId(),
                  fill: "#FFFFFF",
                  content: "万兴剧厂",
                  fontSize: 32,
                  fontWeight: "bold",
                  fontFamily: "Inter"
                }
              ]
            },
            
            // Main Title
            {
              type: "text",
              id: genId(),
              name: "Main Title",
              fill: "#FFFFFF",
              content: "您的专属AI电影工作室",
              fontFamily: "Inter",
              fontSize: 64,
              fontWeight: "bold",
              textAlign: "center"
            },
            
            // Subtitle
            {
              type: "text",
              id: genId(),
              name: "Subtitle",
              fill: "#E0E0E0",
              content: "影视级规模化生产  |  小成本成就大爆款",
              fontFamily: "Inter",
              fontSize: 18,
              fontWeight: "normal",
              textAlign: "center"
            },

            // Create Button
            {
              type: "frame",
              id: genId(),
              name: "Create Button Container",
              layout: "horizontal",
              gap: 0,
              fill: "transparent",
              // we will wrap the button to make it centered-ish or block-level width constraint
              children: [
                {
                  type: "frame",
                  id: genId(),
                  name: "Primary Button",
                  width: 320,
                  height: 64,
                  fill: "#9C42F5", // Purple 
                  cornerRadius: 12,
                  layout: "horizontal",
                  gap: 12,
                  padding: 16,
                  children: [
                    {
                      type: "text",
                      id: genId(),
                      fill: "#FFFFFF",
                      content: "✨ 创建项目",
                      fontSize: 20,
                      fontWeight: "bold",
                      textAlign: "center",
                      width: "fill_container"
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  ],
  themes: {
    Theme: [ "Default" ]
  }
};

const finalJson = {
  version: "2.8",
  children: aiStudioPage.children, // Just the single AI Studio Page as the file's pages/frames
  themes: aiStudioPage.themes
};

fs.writeFileSync('docs/pencil/ai-studio.pen', JSON.stringify(finalJson, null, 2));
console.log('Successfully generated docs/pencil/ai-studio.pen');
