# Mobile Client Style Guide (WeChat Inspired)

## Color Palette
- **Primary Color (WeChat Green):** `#07C160`
- **Background Color:** `#EDEDED` (Light Gray)
- **Card/Item Background:** `#FFFFFF` (White)
- **Primary Text:** `#111111` (Almost Black)
- **Secondary Text:** `#7F7F7F` (Gray)
- **Divider/Border:** `#D9D9D9`
- **Error/Delete:** `#FA5151`

## Typography
- **Font Family:** -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif
- **Headings:** Font-weight 500 or 600
- **Body:** Font-weight 400

## Components

### 1. Navigation Bar (Top)
- **Background:** `#EDEDED` or `#F7F7F7`
- **Text Color:** `#111111`
- **Height:** `44px` (plus status bar)
- **Action Icons:** Black/Dark Gray

### 2. Bottom Tab Bar
- **Background:** `#F7F7F7`
- **Border Top:** `1px solid #D9D9D9`
- **Active Color:** `#07C160`
- **Inactive Color:** `#7F7F7F`
- **Height:** `50px` (plus safe area)

### 3. List Items (Session List)
- **Background:** `#FFFFFF`
- **Padding:** `12px 16px`
- **Border Bottom:** `1px solid #EDEDED` (inset)
- **Title:** 16px, `#111111`
- **Subtitle/Status:** 14px, `#7F7F7F`

### 4. Buttons
- **Primary Button:**
  - Background: `#07C160`
  - Text: `#FFFFFF`
  - Border Radius: `4px`
  - Height: `44px`
- **Secondary/Ghost Button:**
  - Background: Transparent
  - Text: `#576B95` (Link Blue)

### 5. Quick Navigation / Tmux Ops
- **Container:** Floating action button or bottom toolbar within session
- **Style:** Minimalist icons
- **Background:** Semi-transparent or `#F7F7F7`

## Layout Structure
- **Main View (Tab 1):** Session List (Chats)
- **Add Session:** Modal or separate screen
- **Session Detail:** Full screen webview or container for the terminal
