import { COLORS, FormControl, Input, Select } from "@nilfoundation/ui-kit";
import type { FC } from "react";
import { useStyletron } from "styletron-react";
import type { Token } from "./Token";

type TokenInputProps = {
  value: { token: string | Token; amount: string };
  onChange: (value: { token: string | Token; amount: string }) => void;
  tokens: { token: string | Token }[];
  className?: string;
  label?: string;
  disabled?: boolean;
  caption?: string;
};

const TokenInput: FC<TokenInputProps> = ({
  value,
  onChange,
  tokens,
  className,
  label,
  disabled = false,
  caption,
}) => {
  const [css] = useStyletron();

  return (
    <div className={`${css({})} ${className}`}>
      <FormControl label={label} caption={caption}>
        <Input
          disabled={disabled}
          overrides={{
            Input: {
              style: {
                "::-webkit-outer-spin-button": {
                  WebkitAppearance: "none",
                  margin: 0,
                },
                "::-webkit-inner-spin-button": {
                  WebkitAppearance: "none",
                  margin: 0,
                },
                "-moz-appearance": "textfield",
              },
            },
            Root: {
              style: {
                backgroundColor: COLORS.gray700,
                ":hover": {
                  backgroundColor: COLORS.gray600,
                },
              },
            },
          }}
          type="number"
          value={value.amount}
          onChange={(e) => {
            onChange({
              token: value.token,
              amount: e.currentTarget.value,
            });
          }}
          endEnhancer={
            <Select
              disabled={disabled}
              options={tokens.map(({ token }) => ({
                label: token,
                id: token,
              }))}
              searchable={false}
              overrides={{
                ControlContainer: {
                  style: {
                    width: "100px",
                    backgroundColor: "transparent",
                    boxShadow: "none",
                    ":has(input:focus-within)": {
                      boxShadow: "none",
                    },
                    ":hover": {
                      backgroundColor: "transparent",
                    },
                  },
                },
              }}
              placeholder=""
              clearable={false}
              onChange={(params) => {
                onChange({
                  token: params.value[0].label as string,
                  amount: value.amount,
                });
              }}
              value={[
                {
                  label: value.token,
                  id: value.token,
                },
              ]}
            />
          }
        />
      </FormControl>
    </div>
  );
};

export { TokenInput };
