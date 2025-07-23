import Edge from "./Edge";
// import { action } from "@storybook/addon-actions";

const meta = {
  title: "components/Edge",
  component: Edge,
};

export default meta;

export const Default = {
  args: {
    x1: 10,
    y1: 10,
    x2: 110,
    y2: 10,
    active: false,
    userColor: "#007bff",
    onClick: console.log("clicked"),
  },
  render: (args) => (
    <svg width="120" height="20" style={{ border: "1px solid #ccc" }}>
      <Edge {...args} />
    </svg>
  ),
};
