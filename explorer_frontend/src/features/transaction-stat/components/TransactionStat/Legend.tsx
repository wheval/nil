import { HeadingXXLarge } from "baseui/typography";
import { useStyletron } from "styletron-react";
import { styles as s } from "./styles";

type LegendProps = {
  value: string;
};

export const Legend = ({ value }: LegendProps) => {
  const [css] = useStyletron();

  return (
    <div className={css(s.legend)}>
      <HeadingXXLarge>{value}</HeadingXXLarge>
    </div>
  );
};
